import { NextResponse, type NextRequest } from "next/server";

const sessionCookie = "erp_session";

export function proxy(request: NextRequest) {
  if (request.cookies.has(sessionCookie)) {
    return NextResponse.next();
  }

  const login = new URL("/login", request.url);
  login.searchParams.set("next", `${request.nextUrl.pathname}${request.nextUrl.search}`);
  return NextResponse.redirect(login);
}

export const config = {
  matcher: ["/((?!api|login|_next/static|_next/image|favicon.ico|robots.txt).*)"],
};
