import { Button } from "@/components/ui/button";

export default function OrgSettingsMembersPage() {
  return (
    <div className="px-8 py-10 space-y-8">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Settings</p>
          <h1 className="mt-1 text-2xl font-semibold text-foreground">Members</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Manage team members and pending invitations.
          </p>
        </div>
        <Button disabled>Invite Member</Button>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">Active Members</p>
        <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
          <p className="text-sm text-muted-foreground/60">No members yet</p>
        </div>
      </div>

      <div className="rounded-xl border border-border bg-card p-6">
        <p className="text-xs uppercase tracking-[0.2em] text-muted-foreground">
          Pending Invitations
        </p>
        <div className="mt-4 rounded-lg border border-dashed border-border py-10 text-center">
          <p className="text-sm text-muted-foreground/60">No pending invitations</p>
        </div>
      </div>
    </div>
  );
}
