#!/usr/bin/env python3
"""Build and deploy only the template-project Next.js frontend to Cloud Run."""

import argparse
from datetime import datetime, timezone
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
from urllib.request import urlopen

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[3]
FRONTEND = REPO / 'frontend' / 'web'
NAME_RE = re.compile(r'[a-z][a-z0-9-]*')


def config(env):
    required = ['GCP_PROJECT_ID', 'GCP_REGION', 'ARTIFACT_REPOSITORY',
                'BACKEND_SERVICE', 'FRONTEND_SERVICE', 'FRONTEND_RUNTIME_SERVICE_ACCOUNT']
    missing = [key for key in required if not env.get(key, '').strip()]
    if missing:
        raise ValueError('Isi konfigurasi: ' + ', '.join(missing))
    for key in required:
        if not NAME_RE.fullmatch(env[key]):
            raise ValueError(f'{key} harus berupa nama resource yang valid')
    if env['BACKEND_SERVICE'] == env['FRONTEND_SERVICE']:
        raise ValueError('BACKEND_SERVICE dan FRONTEND_SERVICE harus berbeda')
    if env.get('SUB_APP_TYPE', 'admin') not in ('admin', 'customer'):
        raise ValueError('SUB_APP_TYPE harus admin atau customer')
    if env.get('CLOUD_RUN_ALLOW_UNAUTHENTICATED', 'true') not in ('true', 'false'):
        raise ValueError('CLOUD_RUN_ALLOW_UNAUTHENTICATED harus true atau false')
    for key, default in [('CLOUD_RUN_MAX_INSTANCES', '3'), ('CLOUD_RUN_CONCURRENCY', '80')]:
        if not env.get(key, default).isdigit() or int(env.get(key, default)) < 1:
            raise ValueError(f'{key} harus bilangan positif')


def run(project, *args):
    return subprocess.run(['gcloud', *args, '--project', project, '--quiet'],
                          text=True, stdout=subprocess.PIPE, check=True).stdout.strip()


def ensure(project, describe, create):
    try:
        run(project, *describe)
    except subprocess.CalledProcessError:
        run(project, *create)


def deploy(env):
    if not shutil.which('gcloud'):
        raise ValueError('Install Google Cloud CLI dan jalankan gcloud auth login dahulu')
    project, region = env['GCP_PROJECT_ID'], env['GCP_REGION']
    repository, service = env['ARTIFACT_REPOSITORY'], env['FRONTEND_SERVICE']
    account = env['FRONTEND_RUNTIME_SERVICE_ACCOUNT']
    account_email = f'{account}@{project}.iam.gserviceaccount.com'

    run(project, 'services', 'enable', 'run.googleapis.com', 'artifactregistry.googleapis.com',
        'cloudbuild.googleapis.com', 'iam.googleapis.com')
    ensure(project, ['artifacts', 'repositories', 'describe', repository, '--location', region],
           ['artifacts', 'repositories', 'create', repository, '--location', region,
            '--repository-format=docker'])
    ensure(project, ['iam', 'service-accounts', 'describe', account_email],
           ['iam', 'service-accounts', 'create', account])

    try:
        backend_url = run(project, 'run', 'services', 'describe', env['BACKEND_SERVICE'],
                          '--region', region, '--format=value(status.url)')
    except subprocess.CalledProcessError as error:
        raise ValueError('Backend Cloud Run belum tersedia; deploy backend terlebih dahulu.') from error

    sub_app_type = env.get('SUB_APP_TYPE', 'admin')
    configured_backend_url = env.get('BACKEND_BASE_URL', '').strip()
    api_url = configured_backend_url or backend_url
    app_name = env.get('APP_NAME', 'Application').strip() or 'Application'
    tag = datetime.now(timezone.utc).strftime('%Y%m%d-%H%M%S-%f')
    image = f'{region}-docker.pkg.dev/{project}/{repository}/template-project-frontend:{tag}'
    substitutions = ','.join([
        f'_IMAGE={image}',
        f'_SUB_APP_TYPE={sub_app_type}',
        f'_APP_NAME={app_name}',
        f'_NEXT_PUBLIC_ACCOUNT_REGISTER_PATH={env.get("NEXT_PUBLIC_ACCOUNT_REGISTER_PATH", "/auth/accounts/register")}',
        f'_NEXT_PUBLIC_LOGIN_PATH={env.get("NEXT_PUBLIC_LOGIN_PATH", "/auth/login")}',
        f'_NEXT_PUBLIC_FIREBASE_API_KEY={env.get("NEXT_PUBLIC_FIREBASE_API_KEY", "")}',
        f'_NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN={env.get("NEXT_PUBLIC_FIREBASE_AUTH_DOMAIN", "")}',
        f'_NEXT_PUBLIC_FIREBASE_PROJECT_ID={env.get("NEXT_PUBLIC_FIREBASE_PROJECT_ID", "")}',
        f'_NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET={env.get("NEXT_PUBLIC_FIREBASE_STORAGE_BUCKET", "")}',
        f'_NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID={env.get("NEXT_PUBLIC_FIREBASE_MESSAGING_SENDER_ID", "")}',
        f'_NEXT_PUBLIC_FIREBASE_APP_ID={env.get("NEXT_PUBLIC_FIREBASE_APP_ID", "")}',
        f'_NEXT_PUBLIC_FIREBASE_VAPID_KEY={env.get("NEXT_PUBLIC_FIREBASE_VAPID_KEY", "")}',
    ])
    run(project, 'builds', 'submit', str(FRONTEND), '--region', region,
        '--config', str(HERE / 'cloudbuild.yaml'), '--ignore-file', str(HERE / '.gcloudignore'),
        '--substitutions', substitutions)

    args = [
        'run', 'deploy', service, '--image', image, '--region', region,
        '--service-account', account_email, '--execution-environment=gen2', '--port=3000',
        '--cpu=1', '--memory=512Mi', '--min=0',
        '--max-instances', env.get('CLOUD_RUN_MAX_INSTANCES', '3'),
        '--concurrency', env.get('CLOUD_RUN_CONCURRENCY', '80'), '--timeout=120',
    ]
    args.append('--allow-unauthenticated' if env.get('CLOUD_RUN_ALLOW_UNAUTHENTICATED', 'true') == 'true'
                else '--no-allow-unauthenticated')
    args.extend(['--update-env-vars', f'BACKEND_BASE_URL={api_url},APP_NAME={app_name}'])
    run(project, *args)
    url = run(project, 'run', 'services', 'describe', service, '--region', region,
              '--format=value(status.url)')
    print(f'Frontend URL: {url}')
    if env.get('CLOUD_RUN_ALLOW_UNAUTHENTICATED', 'true') == 'true':
        with urlopen(url, timeout=30) as response:
            print(f'Frontend health: HTTP {response.status}')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true', help='Validasi konfigurasi tanpa akses Google Cloud')
    args = parser.parse_args()
    try:
        config(os.environ)
        if args.check:
            print('Konfigurasi frontend Cloud Run valid.')
        else:
            deploy(os.environ)
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        print(f'Deploy frontend gagal: {error}', file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
