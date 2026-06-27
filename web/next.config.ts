import type { NextConfig } from "next";

const skipInternalTypeCheck = process.env.OCTOPUS_NEXT_SKIP_INTERNAL_TS === '1';
const enableWorkerThreads = process.env.OCTOPUS_NEXT_ENABLE_WORKER_THREADS === '1';

const nextConfig: NextConfig = {
  reactCompiler: true,
  output: 'export',
  assetPrefix: './',
  productionBrowserSourceMaps: true,
  experimental: {
    // Worker threads are only enabled when the patched build wrapper opts in.
    // Plain `next build` on Windows can otherwise fail with DataCloneError.
    workerThreads: enableWorkerThreads,
  },
  typescript: {
    // The repo build wrapper runs `tsc --noEmit` first, then opts out of
    // Next's internal type-check worker only for that build invocation.
    ignoreBuildErrors: skipInternalTypeCheck,
  },
};

export default nextConfig;


