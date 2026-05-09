import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

const sourcePath = path.join(repoRoot, 'web/src/components/modules/setting/AIAutomationSource.tsx');
const source = fs.readFileSync(sourcePath, 'utf8');

assert.match(source, /import \{ useAIGovernanceOverview, useStrategyProfiles \} from '@\/api\/endpoints\/ai-automation';/);
assert.match(source, /const overviewQuery = useAIGovernanceOverview\(\);/);
assert.match(source, /const profilesQuery = useStrategyProfiles\(\);/);
assert.match(source, /const setActiveItem = useNavStore\(\(state\) => state\.setActiveItem\);/);
assert.match(source, /const openAICenter = \(\) => setActiveItem\('ai'\);/);
assert.match(source, /data-testid="setting-ai-governance-source"/);
assert.match(source, /const modeLabel = overview\?\.execution_source\.mode === 'manual' \? '手动配置' : overview\?\.execution_source\.mode === 'ai_profile' \? 'AI 策略方案' : '-';/);
assert.match(source, /overview\?\.active_strategy_profile\?\.name \?\? t\('aiAutomationSource\.noActiveProfile'\)/);
assert.match(source, /overview\?\.learning \? `样本 \$\{overview\.learning\.sample_count\}` : '样本 0'/);
assert.match(source, /const recentStatusLabel = overview\?\.recent_session\?\.status \? \(sessionStatusLabelMap\[overview\.recent_session\.status\] \?\? overview\.recent_session\.status\) : '-';/);
assert.match(source, /overview\?\.recent_session\?\.operator_summary \|\| t\('aiAutomationSource\.profileSummaryFallback'\)/);
assert.match(source, /t\('aiAutomationSource\.profileEmpty'\)[\s\S]*strategyProfiles\.length > 0 \? `\$\{strategyProfiles\.length\}` : '0'/);
assert.match(source, /<Button type="button" variant="outline" className="h-10 w-full rounded-xl sm:w-auto sm:min-w-40" onClick=\{openAICenter\}>/);
assert.doesNotMatch(source, /<select/i);
assert.doesNotMatch(source, /role="combobox"/i);

const apiPath = path.join(repoRoot, 'web/src/api/endpoints/ai-automation.ts');
const apiSource = fs.readFileSync(apiPath, 'utf8');
assert.match(apiSource, /export interface AIGovernanceExecutionSource \{/);
assert.match(apiSource, /mode: 'manual' \| 'ai_profile';/);
assert.match(apiSource, /label: string;/);
assert.match(apiSource, /export interface AIGovernanceLearningSummary \{/);
assert.match(apiSource, /sample_count: number;/);
assert.match(apiSource, /export interface StrategyProfileSummary \{/);
assert.match(apiSource, /is_active: boolean;/);
assert.match(apiSource, /export interface AIGovernanceOverview \{/);
assert.match(apiSource, /execution_source: AIGovernanceExecutionSource;/);
assert.match(apiSource, /learning: AIGovernanceLearningSummary;/);
assert.match(apiSource, /active_strategy_profile\?: StrategyProfileSummary;/);
assert.match(apiSource, /recent_session\?: GovernanceSessionSummary;/);
assert.match(apiSource, /export function useAIGovernanceOverview\(\) \{/);
assert.match(apiSource, /export function useStrategyProfiles\(\) \{/);

const testPath = path.join(repoRoot, 'web/src/components/modules/setting/AIAutomationSource.test.tsx');
const testSource = fs.readFileSync(testPath, 'utf8');
assert.match(testSource, /renders governance summary instead of the old profile switcher/);
assert.match(testSource, /queryByRole\('combobox'\)/);
assert.match(testSource, /toBeInTheDocument\(\)/);
assert.match(testSource, /fireEvent\.click\(screen\.getByRole\('button', \{ name: 'aiAutomationSource\.openCenter' \}\)\)/);
assert.match(testSource, /toHaveBeenCalledWith\('ai'\)/);

console.log('ai-config-profile-summary verification passed');
