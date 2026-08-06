import { redirect } from "next/navigation";
import { cookies } from "next/headers";

export default async function Home() {
  const cookieStore = await cookies();
  const hasSession =
    cookieStore.get("session")?.value === "1" ||
    Boolean(cookieStore.get("token")?.value?.trim());

  redirect(hasSession ? "/dashboard" : "/login");
}
