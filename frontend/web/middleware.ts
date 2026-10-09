import { NextResponse, type NextRequest } from 'next/server';
import { resolveSubAppType } from './app/helpers/subApp';

const subAppType = resolveSubAppType(process.env.NEXT_PUBLIC_SUB_APP_TYPE);
const activePrefix = subAppType === 'customer' ? '/customer' : '/admin';
const otherPrefix = subAppType === 'customer' ? '/admin' : '/customer';

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (pathname === otherPrefix || pathname.startsWith(`${otherPrefix}/`)) {
    const destination = request.nextUrl.clone();
    destination.pathname = `${activePrefix}${pathname.slice(otherPrefix.length)}` || activePrefix;
    return NextResponse.redirect(destination);
  }

  if (pathname === activePrefix || pathname.startsWith(`${activePrefix}/`)) {
    const destination = request.nextUrl.clone();
    const appPath = pathname.slice(activePrefix.length) || '/';
    destination.pathname = appPath;
    return NextResponse.rewrite(destination);
  }

  const destination = request.nextUrl.clone();
  destination.pathname = `${activePrefix}${pathname === '/' ? '' : pathname}`;
  return NextResponse.redirect(destination);
}

export const config = {
  matcher: [
    '/((?!api(?:/|$)|_next(?:/|$)|favicon\\.ico$|firebase-messaging-sw\\.js$|.*\\.[^/]+$).*)',
  ],
};
