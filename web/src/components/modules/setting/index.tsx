'use client';

import dynamic from 'next/dynamic';
import { PageWrapper } from '@/components/common/PageWrapper';
import { SettingAppearance } from './Appearance';
import { SettingSystem } from './System';
import { SettingAPIKey } from './APIKey';
import { SettingCostAlert } from './CostAlert';
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
        <div className="octo-workbench">
            <PageWrapper className="space-y-3" disableAnimations>
                <div className="grid grid-cols-1 items-start gap-3 xl:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)]">
                    <div className="min-w-0 space-y-3">
                        <SettingSystem key="setting-system" />
                        <SettingLog key="setting-log" />
                        <SettingAPIKey key="setting-apikey" />
                        <SettingCostAlert key="setting-cost-alert" />
                        <SettingAIAutomationSource key="setting-ai-automation-source" />
                        <LazySettingModelProbe key="setting-model-probe" />
                        <LazySettingBackup key="setting-backup" />
                    </div>
                    <div className="min-w-0 space-y-3">
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
