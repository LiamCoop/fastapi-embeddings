import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";

const isPublicRoute = createRouteMatcher([
  "/sign-in(.*)",
  "/sign-up(.*)",
  "/api/webhooks/clerk(.*)",
]);

export default clerkMiddleware(async (auth, request) => {
  const { userId, sessionClaims } = await auth();
  const { pathname } = request.nextUrl;

  // Unauthenticated users: protect everything except public routes.
  if (!userId) {
    if (!isPublicRoute(request) && pathname !== "/") {
      await auth.protect();
    }
    return NextResponse.next();
  }

  // Authenticated users hitting "/": redirect to their org or /dashboard.
  if (pathname === "/") {
    const orgSlug = (sessionClaims as Record<string, unknown>)?.org_slug as
      | string
      | undefined;
    const dest = orgSlug ? `/org/${orgSlug}` : "/dashboard";
    return NextResponse.redirect(new URL(dest, request.url));
  }

  return NextResponse.next();
});

export const config = {
  matcher: [
    "/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)",
    "/(api|trpc)(.*)",
  ],
};
