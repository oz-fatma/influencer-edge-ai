import { Suspense } from "react";
import LoginForm from "./LoginForm";

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="flex min-h-full items-center justify-center text-[var(--muted)]">
          Loading...
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
