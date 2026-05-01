import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useTranslations } from 'next-intl';
import { CheckCircle2, XCircle, Loader2, Check } from 'lucide-react';
import type { Channel } from '@/api/endpoints/channel';
import { useTestChannelModels } from '@/api/endpoints/channel';

function Checkbox({
    checked,
    onChange,
    id,
}: {
    checked: boolean;
    onChange: (checked: boolean) => void;
    id?: string;
}) {
    return (
        <div
            className="size-4 shrink-0 rounded border border-primary cursor-pointer flex items-center justify-center transition-colors hover:bg-primary/10"
            onClick={() => onChange(!checked)}
        >
            {checked && <Check className="h-3 w-3 text-primary" />}
        </div>
    );
}

interface ModelTabContentProps {
    channel: Channel;
}

export function ModelTabContent({ channel }: ModelTabContentProps) {
    const t = useTranslations('channel.models');
    const testModels = useTestChannelModels();
    const [selectedModels, setSelectedModels] = useState<Set<string>>(new Set());
    const [testResults, setTestResults] = useState<Map<string, { passed: boolean; error?: string; delay?: number }>>(new Map());
    const [isTesting, setIsTesting] = useState(false);

    // 获取所有模型列表（auto + custom）
    const allModels = [
        ...channel.model?.split(',').filter(Boolean) ?? [],
        ...channel.custom_model?.split(',').filter(Boolean) ?? []
    ];

    const handleToggleAll = () => {
        if (selectedModels.size === allModels.length) {
            setSelectedModels(new Set());
        } else {
            setSelectedModels(new Set(allModels));
        }
    };

    const handleTest = async (models?: string[]) => {
        const modelsToTest = models ?? Array.from(selectedModels);
        if (modelsToTest.length === 0) return;

        setIsTesting(true);
        try {
            const results = await testModels.mutateAsync({
                channel_id: channel.id,
                models: modelsToTest,
            });

            // Convert array results to Map
            const resultsMap = new Map<string, { passed: boolean; error?: string; delay?: number }>();
            for (const result of results) {
                resultsMap.set(result.model, {
                    passed: result.passed,
                    error: result.error,
                    delay: result.delay,
                });
            }
            setTestResults(resultsMap);
        } finally {
            setIsTesting(false);
        }
    };

    const handleTestFirst = () => {
        if (allModels.length > 0) {
            handleTest([allModels[0]]);
        }
    };

    return (
        <div className="space-y-4 max-h-[60vh] overflow-y-auto">
            {allModels.length === 0 ? (
                <div className="text-center py-12 text-muted-foreground">
                    {t('noModels')}
                </div>
            ) : (
                <>
                    {/* 工具栏 */}
                    <div className="flex items-center justify-between gap-3">
                        <Button
                            onClick={handleToggleAll}
                            variant="ghost"
                            size="sm"
                            className="h-8"
                        >
                            {selectedModels.size === allModels.length ? t('deselectAll') : t('selectAll')}
                        </Button>
                        <div className="flex items-center gap-2">
                            {selectedModels.size > 0 && (
                                <span className="text-xs text-muted-foreground">
                                    {t('selectedCount', { count: selectedModels.size })}
                                </span>
                            )}
                            <Button
                                onClick={() => selectedModels.size > 0 ? handleTest() : handleTestFirst()}
                                disabled={isTesting}
                                size="sm"
                                className="h-8"
                            >
                                {isTesting ? (
                                    <>
                                        <Loader2 className="h-4 w-4 mr-1 animate-spin" />
                                        {t('testing')}
                                    </>
                                ) : selectedModels.size > 0 ? (
                                    t('testSelected')
                                ) : (
                                    t('test')
                                )}
                            </Button>
                        </div>
                    </div>

                    {/* 模型列表 */}
                    <div className="space-y-2">
                        {allModels.map((model) => {
                            const isSelected = selectedModels.has(model);
                            const result = testResults.get(model);

                            return (
                                <div
                                    key={model}
                                    className="flex items-center gap-3 p-3 border rounded-xl hover:bg-accent/5 transition-colors"
                                >
                                    <Checkbox
                                        id={`model-${model}`}
                                        checked={isSelected}
                                        onChange={(checked) => {
                                            const newSet = new Set(selectedModels);
                                            if (checked) newSet.add(model);
                                            else newSet.delete(model);
                                            setSelectedModels(newSet);
                                        }}
                                    />
                                    <label
                                        htmlFor={`model-${model}`}
                                        className="flex-1 font-mono text-sm cursor-pointer select-all"
                                    >
                                        {model}
                                    </label>
                                    {result && (
                                        <div className="flex items-center gap-2">
                                            {result.delay !== undefined && (
                                                <span className="text-xs text-muted-foreground">
                                                    {result.delay}ms
                                                </span>
                                            )}
                                            <Badge
                                                variant="secondary"
                                                className={result.passed ? 'bg-green-500/15 text-green-700 dark:text-green-400' : 'bg-red-500/15 text-red-700 dark:text-red-400'}
                                            >
                                                {result.passed ? (
                                                    <>
                                                        <CheckCircle2 className="h-3 w-3 mr-1" />
                                                        {t('testPassed')}
                                                    </>
                                                ) : (
                                                    <>
                                                        <XCircle className="h-3 w-3 mr-1" />
                                                        {t('testFailed')}
                                                    </>
                                                )}
                                            </Badge>
                                        </div>
                                    )}
                                </div>
                            );
                        })}
                    </div>

                    {/* 测试结果摘要 */}
                    {testResults.size > 0 && (
                        <div className="text-xs text-muted-foreground">
                            {t('testSummary', {
                                total: testResults.size,
                                passed: Array.from(testResults.values()).filter((r) => r.passed).length,
                            })}
                        </div>
                    )}
                </>
            )}
        </div>
    );
}
