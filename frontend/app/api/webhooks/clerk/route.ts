import { headers } from "next/headers";
import { Webhook } from "svix";
import type { WebhookEvent } from "@clerk/nextjs/server";
import { getDb } from "@/lib/db";
import { User } from "@/lib/entities/user.entity";
import { Organization } from "@/lib/entities/organization.entity";
import { OrganizationMembership } from "@/lib/entities/organization-membership.entity";
import { generateOrgSlug } from "@/lib/org-slug";

function getHeaderValue(headerMap: Headers, key: string) {
  const value = headerMap.get(key);
  return value ?? "";
}

export async function POST(req: Request) {
  const webhookSecret = process.env.CLERK_WEBHOOK_SECRET;

  if (!webhookSecret) {
    return new Response("Missing CLERK_WEBHOOK_SECRET", { status: 500 });
  }

  const payload = await req.text();
  const headerList = await headers();

  const svixId = getHeaderValue(headerList, "svix-id");
  const svixTimestamp = getHeaderValue(headerList, "svix-timestamp");
  const svixSignature = getHeaderValue(headerList, "svix-signature");

  if (!svixId || !svixTimestamp || !svixSignature) {
    return new Response("Missing svix headers", { status: 400 });
  }

  const wh = new Webhook(webhookSecret);
  let event: WebhookEvent;

  try {
    event = wh.verify(payload, {
      "svix-id": svixId,
      "svix-timestamp": svixTimestamp,
      "svix-signature": svixSignature,
    }) as WebhookEvent;
  } catch (error) {
    console.error("Clerk webhook verification failed", error);
    return new Response("Invalid signature", { status: 400 });
  }

  const db = await getDb();
  const userRepo = db.getRepository(User);
  const orgRepo = db.getRepository(Organization);
  const membershipRepo = db.getRepository(OrganizationMembership);

  if (event.type === "user.created") {
    const primaryEmail = event.data.email_addresses.find(
      (e) => e.id === event.data.primary_email_address_id
    );
    const email = primaryEmail?.email_address ?? "";

    const user = userRepo.create({
      clerkId: event.data.id,
      email,
      firstName: event.data.first_name,
      lastName: event.data.last_name,
    });
    await userRepo.save(user);

    const slug = generateOrgSlug(
      event.data.first_name,
      event.data.last_name,
      email,
      event.data.id
    );

    const org = orgRepo.create({ name: slug, slug });
    await orgRepo.save(org);

    const membership = membershipRepo.create({
      organizationId: org.id,
      userClerkId: event.data.id,
      role: "owner",
      acceptedAt: new Date(),
    });
    await membershipRepo.save(membership);

    console.info("User created with personal org", { userId: event.data.id, slug });
  }

  if (event.type === "user.updated") {
    const primaryEmail = event.data.email_addresses.find(
      (e) => e.id === event.data.primary_email_address_id
    );
    await userRepo.update(
      { clerkId: event.data.id },
      {
        email: primaryEmail?.email_address ?? "",
        firstName: event.data.first_name,
        lastName: event.data.last_name,
      }
    );
    console.info("User updated", { userId: event.data.id });
  }

  if (event.type === "user.deleted") {
    await userRepo.delete({ clerkId: event.data.id });
    console.info("User deleted", { userId: event.data.id });
  }

  return new Response("ok", { status: 200 });
}
