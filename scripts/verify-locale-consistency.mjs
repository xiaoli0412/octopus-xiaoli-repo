import assert from 'node:assert/strict';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, '..');

function loadLocale(name) {
  const filePath = path.join(repoRoot, `web/public/locale/${name}.json`);
  return JSON.parse(fs.readFileSync(filePath, 'utf8'));
}

function pickPaths(root, paths) {
  return paths.flatMap((keyPath) => {
    const parts = keyPath.split('.');
    let cursor = root;
    for (const part of parts) {
      if (cursor && typeof cursor === 'object' && part in cursor) {
        cursor = cursor[part];
      } else {
        return [];
      }
    }
    return [{ keyPath, value: cursor }];
  });
}

function collectStrings(node, prefix = '') {
  if (!node || typeof node !== 'object' || Array.isArray(node)) {
    return typeof node === 'string' ? [{ keyPath: prefix, value: node }] : [];
  }
  return Object.entries(node).flatMap(([key, value]) => {
    const nextPrefix = prefix ? `${prefix}.${key}` : key;
    return collectStrings(value, nextPrefix);
  });
}

const ja = loadLocale('ja');
const zhHans = loadLocale('zh-Hans');
const zhHant = loadLocale('zh-Hant');
const en = loadLocale('en');

const keyModules = [
  'apiKeyDashboard',
  'aiAutomation',
  'navbar',
  'channel.card',
  'channel.detail',
  'channel.form',
  'group.create',
  'group.form',
  'home',
  'setting.circuitBreaker',
  'setting.dynamicRouting',
  'setting.backup',
  'setting.apiKey',
  'setting.llmPrice',
  'setting.modelProbe',
  'setting.llmSync',
  'setting.info',
  'doc',
];

const jaStrings = pickPaths(ja, keyModules).flatMap(({ keyPath, value }) => collectStrings(value, keyPath));
const localizedUiModules = ['aiAutomation', 'channel.card', 'channel.detail', 'channel.form', 'group.form', 'home', 'setting', 'doc'];
const zhHansStrings = pickPaths(zhHans, localizedUiModules).flatMap(({ keyPath, value }) => collectStrings(value, keyPath));
const zhHantStrings = pickPaths(zhHant, localizedUiModules).flatMap(({ keyPath, value }) => collectStrings(value, keyPath));
const enStrings = pickPaths(en, localizedUiModules).flatMap(({ keyPath, value }) => collectStrings(value, keyPath));

const jaBlockedPatterns = [
  { re: /没有匹配的キー/, why: 'mixed Chinese in Japanese key empty state' },
  { re: /每个新请求/, why: 'mixed Chinese routing description' },
  { re: /创建|删除|设置|导入|导出|恢复|预检|备份/, why: 'remaining simplified Chinese workflow copy' },
  { re: /计费モード/, why: 'mixed Chinese or malformed billing label in Japanese locale' },
  { re: /チャネル同步|今すぐ同步|上次同步/, why: 'mixed Chinese/Japanese sync copy' },
  { re: /OpenAI Chat 格式|OpenAI Responses 格式|Anthropic Messages 格式|Embedding 格式/, why: 'Chinese suffix leaked into Japanese docs' },
  { re: /加载|风险|关键词/, why: 'remaining simplified Chinese help or error text' },
  { re: /総 Token|入力 Token|出力 Token|Token 使用量/, why: 'English token wording leaked into Japanese locale' },
  { re: /\bProfile\b|\bendpoint\b|\bbase URL\b|group_items|\bpriority\b|\bmanual\b|OpenAI-compatible/, why: 'English workflow wording leaked into Japanese locale' },
];

const zhBlockedPatterns = [
  { re: /\bRound Robin\b|\bFill Priority\b|\bPriority Order\b|\bHalf-Open\b|\bSettings\b/, why: 'English leakage in Chinese locales' },
  { re: /\bKey\b/, why: 'English key wording leaked into Chinese locales' },
  { re: /\bProfile\b|\bHybrid\b|\bShadow AI\b|\bendpoint\b|\bbase URL\b|API Key|group_items|\bpriority\b|\bmanual\b|OpenAI-compatible/, why: 'English workflow wording leaked into Chinese locales' },
];

const enBlockedPatterns = [
  { re: /创建|删除|设置|备份|恢复|导入|导出|探测|熔断|渠道|分组|密钥/, why: 'Chinese leakage in English locale' },
];

for (const { keyPath, value } of jaStrings) {
  for (const { re, why } of jaBlockedPatterns) {
    assert.ok(!re.test(value), `ja locale leakage at ${keyPath}: ${why} -> ${value}`);
  }
}

for (const { keyPath, value } of zhHansStrings) {
  for (const { re, why } of zhBlockedPatterns) {
    assert.ok(!re.test(value), `zh-Hans locale leakage at ${keyPath}: ${why} -> ${value}`);
  }
}

for (const { keyPath, value } of zhHantStrings) {
  for (const { re, why } of zhBlockedPatterns) {
    assert.ok(!re.test(value), `zh-Hant locale leakage at ${keyPath}: ${why} -> ${value}`);
  }
}

for (const { keyPath, value } of enStrings) {
  for (const { re, why } of enBlockedPatterns) {
    assert.ok(!re.test(value), `en locale leakage at ${keyPath}: ${why} -> ${value}`);
  }
}

assert.equal(ja.setting.language.ja, '日本語');
assert.equal(ja.setting.language.en, '英語');
assert.equal(ja.setting.llmRouteTarget.billingMode, '課金モード');
assert.equal(ja.doc.endpointOpenAIChat, 'OpenAI Chat 形式');
assert.equal(ja.home.tokenBreakdown.total, '合計');

assert.equal(zhHans.channel.card.keyCountBadge, '密钥 {count}');
assert.equal(zhHans.channel.detail.labels.maskedKey, '密钥预览');
assert.equal(zhHans.channel.detail.metrics.totalToken, '总令牌');
assert.equal(zhHans.setting.dynamicRouting.modeOptions['shadow-ai'], '影子模式');
assert.equal(zhHans.setting.dynamicRouting.modeOptions.hybrid, '混合模式');

assert.equal(zhHant.channel.card.keyCountBadge, '金鑰 {count}');
assert.equal(zhHant.channel.detail.labels.enabledState, '啟用狀態');
assert.equal(zhHant.channel.detail.metrics.totalToken, '總令牌');
assert.equal(zhHant.setting.dynamicRouting.modeOptions['shadow-ai'], '影子模式');
assert.equal(zhHant.setting.dynamicRouting.modeOptions.hybrid, '混合模式');

assert.equal(en.channel.card.keyCountBadge, 'Keys {count}');
assert.equal(en.channel.detail.labels.maskedKey, 'Key preview');
assert.equal(en.channel.detail.statusBadge.notChecked, 'Not checked');

assert.equal(ja.channel.card.keyCountBadge, 'キー {count}');
assert.equal(ja.channel.detail.labels.maskedKey, 'キープレビュー');
assert.equal(ja.channel.detail.metrics.totalToken, '総トークン');

assert.equal(zhHans.home.tokenBreakdown.title, '令牌明细');
assert.equal(zhHans.home.tokenBreakdown.estimatedGateway, '估算网关价格');
assert.equal(zhHans.home.total.inputTokens, '输入令牌');
assert.equal(zhHant.home.tokenBreakdown.title, '令牌明細');
assert.equal(zhHant.home.tokenBreakdown.estimatedGateway, '估算網關價格');
assert.equal(ja.home.tokenBreakdown.title, 'トークン明細');
assert.equal(ja.home.tokenBreakdown.estimatedGateway, '推定ゲートウェイ価格');

console.log('locale-consistency verification passed');
