export function generateOrgSlug(
  firstName: string | null,
  lastName: string | null,
  email: string,
  clerkId: string
): string {
  const namePart = [firstName, lastName]
    .filter(Boolean)
    .join("-")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");

  const emailPrefix = email
    .split("@")[0]
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-|-$/g, "");

  const base = namePart || emailPrefix || "org";
  const suffix = clerkId.slice(-6).toLowerCase().replace(/[^a-z0-9]/g, "");

  return `${base}-${suffix}`;
}
