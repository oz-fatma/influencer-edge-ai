"use client";

import {
  createContext,
  useContext,
  useMemo,
  useState,
  type Dispatch,
  type ReactNode,
  type SetStateAction,
} from "react";
import type { InfluencerAnalysisResult } from "@/lib/api";

export type AnalysisPhase = "idle" | "server" | "browser";
export type AnalysisSource = "ollama" | "web-llm" | null;

type AnalysisContextValue = {
  analyzing: boolean;
  setAnalyzing: Dispatch<SetStateAction<boolean>>;
  analysisPhase: AnalysisPhase;
  setAnalysisPhase: Dispatch<SetStateAction<AnalysisPhase>>;
  selectedInfluencerId: string | null;
  setSelectedInfluencerId: Dispatch<SetStateAction<string | null>>;
  liveResult: InfluencerAnalysisResult | null;
  setLiveResult: Dispatch<SetStateAction<InfluencerAnalysisResult | null>>;
  analysisSource: AnalysisSource;
  setAnalysisSource: Dispatch<SetStateAction<AnalysisSource>>;
  analysisError: string | null;
  setAnalysisError: Dispatch<SetStateAction<string | null>>;
  analysisNotice: string | null;
  setAnalysisNotice: Dispatch<SetStateAction<string | null>>;
};

const AnalysisContext = createContext<AnalysisContextValue | null>(null);

export function AnalysisProvider({ children }: { children: ReactNode }) {
  const [analyzing, setAnalyzing] = useState(false);
  const [analysisPhase, setAnalysisPhase] = useState<AnalysisPhase>("idle");
  const [selectedInfluencerId, setSelectedInfluencerId] = useState<string | null>(null);
  const [liveResult, setLiveResult] = useState<InfluencerAnalysisResult | null>(null);
  const [analysisSource, setAnalysisSource] = useState<AnalysisSource>(null);
  const [analysisError, setAnalysisError] = useState<string | null>(null);
  const [analysisNotice, setAnalysisNotice] = useState<string | null>(null);

  const value = useMemo(
    () => ({
      analyzing,
      setAnalyzing,
      analysisPhase,
      setAnalysisPhase,
      selectedInfluencerId,
      setSelectedInfluencerId,
      liveResult,
      setLiveResult,
      analysisSource,
      setAnalysisSource,
      analysisError,
      setAnalysisError,
      analysisNotice,
      setAnalysisNotice,
    }),
    [
      analyzing,
      analysisPhase,
      selectedInfluencerId,
      liveResult,
      analysisSource,
      analysisError,
      analysisNotice,
    ],
  );

  return (
    <AnalysisContext.Provider value={value}>{children}</AnalysisContext.Provider>
  );
}

export function useAnalysisContext(): AnalysisContextValue {
  const ctx = useContext(AnalysisContext);
  if (!ctx) {
    throw new Error("useAnalysisContext must be used within AnalysisProvider");
  }
  return ctx;
}
