'use client';

import { useState } from 'react';
import { BILLING_MODE_OPTIONS, useCreateModel } from '@/api/endpoints/model';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Button } from '@/components/ui/button';
import { Field, FieldLabel, FieldGroup } from '@/components/ui/field';
import {
    MorphingDialogClose,
    MorphingDialogTitle,
    MorphingDialogDescription,
    useMorphingDialog,
} from '@/components/ui/morphing-dialog';
import { useTranslations } from 'next-intl';
import { getBillingModeKey } from '@/lib/ui-labels';

export function CreateDialogContent() {
    const { setIsOpen } = useMorphingDialog();
    const t = useTranslations('model.create');
    const tSetting = useTranslations('setting.llmRouteTarget');
    const createModel = useCreateModel();

    const [formData, setFormData] = useState({
        name: '',
        canonical_name: '',
        input: '',
        output: '',
        cache_read: '',
        cache_write: '',
        official_input: '',
        official_output: '',
        official_cache_read: '',
        official_cache_write: '',
        billing_mode: 'unknown',
    });

    const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
        event.preventDefault();
        if (!formData.name.trim()) return;

        createModel.mutate({
            name: formData.name.trim(),
            canonical_name: formData.canonical_name.trim(),
            input: parseFloat(formData.input) || 0,
            output: parseFloat(formData.output) || 0,
            cache_read: parseFloat(formData.cache_read) || 0,
            cache_write: parseFloat(formData.cache_write) || 0,
            official_input: parseFloat(formData.official_input) || 0,
            official_output: parseFloat(formData.official_output) || 0,
            official_cache_read: parseFloat(formData.official_cache_read) || 0,
            official_cache_write: parseFloat(formData.official_cache_write) || 0,
            billing_mode: formData.billing_mode as 'unknown' | 'per_request' | 'per_token' | 'per_quota' | 'flat' | 'free',
        }, {
            onSuccess: () => {
                setFormData({ name: '', canonical_name: '', input: '', output: '', cache_read: '', cache_write: '', official_input: '', official_output: '', official_cache_read: '', official_cache_write: '', billing_mode: 'unknown' });
                setIsOpen(false);
            }
        });
    };

    return (
        <div className="w-screen max-w-full md:max-w-xl">
            <MorphingDialogTitle>
                <header className="mb-5 flex items-center justify-between">
                    <h2 className="text-2xl font-bold text-card-foreground">{t('title')}</h2>
                    <MorphingDialogClose
                        className="relative right-0 top-0"
                        variants={{
                            initial: { opacity: 0, scale: 0.8 },
                            animate: { opacity: 1, scale: 1 },
                            exit: { opacity: 0, scale: 0.8 },
                        }}
                    />
                </header>
            </MorphingDialogTitle>
            <MorphingDialogDescription>
                <form onSubmit={handleSubmit}>
                    <FieldGroup className="gap-4">
                        <Field>
                            <FieldLabel htmlFor="model-name">{t('name')}</FieldLabel>
                            <Input
                                id="model-name"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                className="rounded-xl"
                            />
                        </Field>
                        <Field>
                            <FieldLabel htmlFor="model-canonical-name">{tSetting('canonicalName')}</FieldLabel>
                            <Input
                                id="model-canonical-name"
                                value={formData.canonical_name}
                                onChange={(e) => setFormData({ ...formData, canonical_name: e.target.value })}
                                className="rounded-xl"
                            />
                        </Field>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <Field>
                                <FieldLabel htmlFor="model-input">{t('input')}</FieldLabel>
                                <Input
                                    id="model-input"
                                    type="number"
                                    step="any"
                                    value={formData.input}
                                    onChange={(e) => setFormData({ ...formData, input: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-output">{t('output')}</FieldLabel>
                                <Input
                                    id="model-output"
                                    type="number"
                                    step="any"
                                    value={formData.output}
                                    onChange={(e) => setFormData({ ...formData, output: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-cache-read">{t('cacheRead')}</FieldLabel>
                                <Input
                                    id="model-cache-read"
                                    type="number"
                                    step="any"
                                    value={formData.cache_read}
                                    onChange={(e) => setFormData({ ...formData, cache_read: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-cache-write">{t('cacheWrite')}</FieldLabel>
                                <Input
                                    id="model-cache-write"
                                    type="number"
                                    step="any"
                                    value={formData.cache_write}
                                    onChange={(e) => setFormData({ ...formData, cache_write: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-official-input">{tSetting('officialInput')}</FieldLabel>
                                <Input
                                    id="model-official-input"
                                    type="number"
                                    step="any"
                                    value={formData.official_input}
                                    onChange={(e) => setFormData({ ...formData, official_input: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-official-output">{tSetting('officialOutput')}</FieldLabel>
                                <Input
                                    id="model-official-output"
                                    type="number"
                                    step="any"
                                    value={formData.official_output}
                                    onChange={(e) => setFormData({ ...formData, official_output: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-official-cache-read">{tSetting('officialCacheRead')}</FieldLabel>
                                <Input
                                    id="model-official-cache-read"
                                    type="number"
                                    step="any"
                                    value={formData.official_cache_read}
                                    onChange={(e) => setFormData({ ...formData, official_cache_read: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-official-cache-write">{tSetting('officialCacheWrite')}</FieldLabel>
                                <Input
                                    id="model-official-cache-write"
                                    type="number"
                                    step="any"
                                    value={formData.official_cache_write}
                                    onChange={(e) => setFormData({ ...formData, official_cache_write: e.target.value })}
                                    className="rounded-xl"
                                />
                            </Field>
                            <Field>
                                <FieldLabel htmlFor="model-billing-mode">{tSetting('billingMode')}</FieldLabel>
                                <Select value={formData.billing_mode} onValueChange={(value) => setFormData({ ...formData, billing_mode: value as typeof formData.billing_mode })}>
                                    <SelectTrigger id="model-billing-mode" className="rounded-xl">
                                        <SelectValue />
                                    </SelectTrigger>
                                    <SelectContent className='rounded-xl'>
                                        {BILLING_MODE_OPTIONS.map((mode) => (
                                            <SelectItem key={mode} className='rounded-xl' value={mode}>{tSetting(`billingModeOptions.${getBillingModeKey(mode)}`)}</SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                            </Field>
                        </div>
                        <Button
                            type="submit"
                            disabled={createModel.isPending || !formData.name.trim()}
                            className="w-full rounded-xl h-11"
                        >
                            {createModel.isPending ? t('submitting') : t('submit')}
                        </Button>
                    </FieldGroup>
                </form>
            </MorphingDialogDescription>
        </div>
    );
}
