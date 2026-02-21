import { auth, currentUser, clerkClient } from "@clerk/nextjs/server";
import { NextResponse } from "next/server";
import { getDb } from "@/lib/db";
import { Organization } from "@/lib/entities/organization.entity";
import { OrganizationMembership } from "@/lib/entities/organization-membership.entity";
import { generateOrgSlug } from "@/lib/org-slug";

export async function GET(request: Request) {
  const { userId, sessionClaims } = await auth();

  if (!userId) {
    return NextResponse.redirect(new URL("/sign-in", request.url));
  }

  // Fast path: slug already in the session token.
  const existingSlug = (sessionClaims as Record<string, unknown>)?.org_slug as
    | string
    | undefined;
  if (existingSlug) {
    return NextResponse.redirect(new URL(`/org/${existingSlug}`, request.url));
  }

  // First login: create the personal org.
  try {
    const clerkUser = await currentUser();
    const email =
      clerkUser?.emailAddresses.find((e) => e.id === clerkUser.primaryEmailAddressId)
        ?.emailAddress ?? "";

    const slug = generateOrgSlug(
      clerkUser?.firstName ?? null,
      clerkUser?.lastName ?? null,
      email,
      userId
    );

    const db = await getDb();
    const orgRepo = db.getRepository(Organization);
    const membershipRepo = db.getRepository(OrganizationMembership);

    let org = await orgRepo.findOne({ where: { slug } });
    if (!org) {
      org = orgRepo.create({ name: slug, slug });
      await orgRepo.save(org);
    }

    const existing = await membershipRepo.findOne({
      where: { userClerkId: userId, organizationId: org.id },
    });
    if (!existing) {
      const membership = membershipRepo.create({
        organizationId: org.id,
        userClerkId: userId,
        role: "owner",
        acceptedAt: new Date(),
      });
      await membershipRepo.save(membership);
    }

    // Write to Clerk metadata in the background so future logins hit the fast path.
    clerkClient()
      .then((client) =>
        client.users.updateUserMetadata(userId, { publicMetadata: { org_slug: slug } })
      )
      .catch((err) => console.error("[dashboard] Failed to write org_slug to Clerk:", err));

    return NextResponse.redirect(new URL(`/org/${slug}`, request.url));
  } catch (err) {
    console.error("[dashboard] Failed to create org for user", userId, err);
    return NextResponse.redirect(new URL("/sign-in", request.url));
  }
}
