#!/usr/bin/env python3
"""Build and deploy only the template-project Go backend to Cloud Run."""

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
BACKEND = REPO / 'backend'
NAME_RE = re.compile(r'[a-z][a-z0-9-]*')


def config(env):
    required = ['GCP_PROJECT_ID', 'GCP_REGION', 'ARTIFACT_REPOSITORY',
                'BACKEND_SERVICE', 'BACKEND_RUNTIME_SERVICE_ACCOUNT']
    missing = [key for key in required if not env.get(key, '').strip()]
    if missing:
        raise ValueError('Isi konfigurasi: ' + ', '.join(missing))
    for key in required:
        if not NAME_RE.fullmatch(env[key]):
            raise ValueError(f'{key} harus berupa nama resource yang valid')
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
    repository, service = env['ARTIFACT_REPOSITORY'], env['BACKEND_SERVICE']
    account = env['BACKEND_RUNTIME_SERVICE_ACCOUNT']
    account_email = f'{account}@{project}.iam.gserviceaccount.com'

    run(project, 'services', 'enable', 'run.googleapis.com', 'artifactregistry.googleapis.com',
        'cloudbuild.googleapis.com', 'iam.googleapis.com')
    ensure(project, ['artifacts', 'repositories', 'describe', repository, '--location', region],
           ['artifacts', 'repositories', 'create', repository, '--location', region,
            '--repository-format=docker'])
    ensure(project, ['iam', 'service-accounts', 'describe', account_email],
           ['iam', 'service-accounts', 'create', account])

    tag = datetime.now(timezone.utc).strftime('%Y%m%d-%H%M%S-%f')
    image = f'{region}-docker.pkg.dev/{project}/{repository}/template-project-backend:{tag}'
    run(project, 'builds', 'submit', str(BACKEND), '--region', region,
        '--config', str(HERE / 'cloudbuild.yaml'), '--ignore-file', str(HERE / '.gcloudignore'),
        f'--substitutions=_IMAGE={image}')

    args = [
        'run', 'deploy', service, '--image', image, '--region', region,
        '--service-account', account_email, '--execution-environment=gen2', '--port=8080',
        '--cpu=1', '--memory=256Mi', '--min=0',
        '--max-instances', env.get('CLOUD_RUN_MAX_INSTANCES', '3'),
        '--concurrency', env.get('CLOUD_RUN_CONCURRENCY', '80'), '--timeout=120',
        '--set-env-vars', f'CORS_ALLOW_ORIGIN={env.get("BACKEND_CORS_ALLOW_ORIGIN", "*")}',
    ]
    args.append('--allow-unauthenticated' if env.get('CLOUD_RUN_ALLOW_UNAUTHENTICATED', 'true') == 'true'
                else '--no-allow-unauthenticated')
    run(project, *args)
    url = run(project, 'run', 'services', 'describe', service, '--region', region,
              '--format=value(status.url)')
    print(f'Backend URL: {url}')
    if env.get('CLOUD_RUN_ALLOW_UNAUTHENTICATED', 'true') == 'true':
        with urlopen(f'{url}/healthz', timeout=30) as response:
            print(f'Backend health: HTTP {response.status}')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true', help='Validasi konfigurasi tanpa akses Google Cloud')
    args = parser.parse_args()
    try:
        config(os.environ)
        if args.check:
            print('Konfigurasi backend Cloud Run valid.')
        else:
            deploy(os.environ)
    except (ValueError, OSError, subprocess.CalledProcessError) as error:
        print(f'Deploy backend gagal: {error}', file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
