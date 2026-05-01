#!/usr/bin/env node

const { spawnSync } = require('node:child_process')
const path = require('node:path')
const { existsSync } = require('node:fs')

const repoRoot = path.resolve(__dirname, '..')
const webDir = path.join(repoRoot, 'web')
const nextBin = path.join(
  webDir,
  'node_modules',
  '.bin',
  process.platform === 'win32' ? 'next.cmd' : 'next',
)

if (!existsSync(nextBin)) {
  console.error('Failed to locate Next.js binary. Run pnpm install in web/.')
  process.exit(1)
}

const args = ['dev', '--experimental-https']

const result = spawnSync(nextBin, args, {
  cwd: webDir,
  stdio: 'inherit',
  shell: process.platform === 'win32',
})

if (result.error) {
  console.error(`Failed to start Next.js dev server: ${result.error.message}`)
  process.exit(1)
}

if (typeof result.signal === 'string' && result.signal.length > 0) {
  console.error(`Next.js dev server exited via signal ${result.signal}`)
  process.exit(1)
}

process.exit(result.status || 0)
