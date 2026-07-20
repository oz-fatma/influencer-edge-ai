export const platformColors: Record<string, string> = {
  instagram: "bg-pink-500/15 text-pink-400",
  tiktok: "bg-cyan-500/15 text-cyan-400",
  youtube: "bg-red-500/15 text-red-400",
  linkedin: "bg-blue-500/15 text-blue-400",
  twitter: "bg-sky-500/15 text-sky-400",
};

export function scoreColor(score: number) {
  if (score >= 85) return "text-[var(--success)]";
  if (score >= 75) return "text-[var(--accent)]";
  return "text-[var(--warning)]";
}
