'use client';

import dynamic from 'next/dynamic';
import { PageWrapper } from '@/components/common/PageWrapper';
import { SettingAppearance } from './Appearance';
import { SettingSystem } from './System';
import { SettingAPIKey } from './APIKey';
import { SettingAccount } from './Account';
import { SettingInfo } from './Info';
import { SettingLLMSync } from './LLMSync';
import { SettingLog } from './Log';
import { SettingAIAutomationSource } from './AIAutomationSource';

const LazySettingLLMPrice = dynamic(() => import('./LLMPrice').then((mod) => ({ default: mod.SettingLLMPrice })));
const LazySettingModelProbe = dynamic(() => import('./ModelProbe').then((mod) => ({ default: mod.SettingModelProbe })));
const LazySettingCircuitBreaker = dynamic(() => import('./CircuitBreaker').then((mod) => ({ default: mod.SettingCircuitBreaker })));
const LazySettingDynamicRouting = dynamic(() => import('./DynamicRouting').then((mod) => ({ default: mod.SettingDynamicRouting })));
const LazySettingBackup = dynamic(() => import('./Backup').then((mod) => ({ default: mod.SettingBackup })));

export function Setting() {
    return (
        <div className="h-full min-h-0 overflow-y-auto overscroll-contain rounded-t-3xl">
            <PageWrapper className="pb-24 md:pb-4">
                <div className="grid grid-cols-1 items-start gap-4 xl:grid-cols-2">
                    <div className="min-w-0 space-y-4">
                        <SettingSystem key="setting-system" />
                        <SettingLog key="setting-log" />
                        <SettingAPIKey key="setting-apikey" />
                        <SettingAIAutomationSource key="setting-ai-automation-source" />
                        <LazySettingModelProbe key="setting-model-probe" />
                        <LazySettingBackup key="setting-backup" />
                    </div>
                    <div className="min-w-0 space-y-4">
                        <SettingInfo key="setting-info" />
                        <SettingAppearance key="setting-appearance" />
                        <SettingAccount key="setting-account" />
                        <LazySettingLLMPrice key="setting-llmprice" />
                        <SettingLLMSync key="setting-llmsync" />
                        <LazySettingCircuitBreaker key="setting-circuit-breaker" />
                        <LazySettingDynamicRouting key="setting-dynamic-routing" />
                    </div>
                </div>
            </PageWrapper>
        </div>
    );
}
