import { Suspense } from "react";
import LoginForm from "./LoginForm";
import { LoginLoadingFallback } from "./LoginLoadingFallback";

export default function LoginPage() {
  return (
    <Suspense fallback={<LoginLoadingFallback />}>
      <LoginForm />
    </Suspense>
  );
}
