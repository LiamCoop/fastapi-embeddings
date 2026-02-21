import { SignUp } from "@clerk/nextjs";

export default function SignUpPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-950 px-6 py-16">
      <SignUp forceRedirectUrl="/" />
    </div>
  );
}
