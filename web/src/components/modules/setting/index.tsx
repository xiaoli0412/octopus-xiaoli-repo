'use client';

import dynamic from 'next/dynamic';
import { type ReactNode } from 'react';
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

function SettingMasonryItem({ children }: { children: ReactNode }) {
    return <div className="mb-4 break-inside-avoid">{children}</div>;
}

export function Setting() {
    return (
        <div className="h-full min-h-0 overflow-y-auto overscroll-contain rounded-t-3xl">
            <PageWrapper className="columns-1 gap-4 pb-24 md:columns-2 md:pb-4">
                <SettingMasonryItem><SettingInfo key="setting-info" /></SettingMasonryItem>
                <SettingMasonryItem><SettingAppearance key="setting-appearance" /></SettingMasonryItem>
                <SettingMasonryItem><SettingAccount key="setting-account" /></SettingMasonryItem>
                <SettingMasonryItem><SettingSystem key="setting-system" /></SettingMasonryItem>
                <SettingMasonryItem><SettingLog key="setting-log" /></SettingMasonryItem>
                <SettingMasonryItem><SettingAPIKey key="setting-apikey" /></SettingMasonryItem>
                <SettingMasonryItem><LazySettingLLMPrice key="setting-llmprice" /></SettingMasonryItem>
                <SettingMasonryItem><SettingAIAutomationSource key="setting-ai-automation-source" /></SettingMasonryItem>
                <SettingMasonryItem><LazySettingModelProbe key="setting-model-probe" /></SettingMasonryItem>
                <SettingMasonryItem><SettingLLMSync key="setting-llmsync" /></SettingMasonryItem>
                <SettingMasonryItem><LazySettingCircuitBreaker key="setting-circuit-breaker" /></SettingMasonryItem>
                <SettingMasonryItem><LazySettingDynamicRouting key="setting-dynamic-routing" /></SettingMasonryItem>
                <SettingMasonryItem><LazySettingBackup key="setting-backup" /></SettingMasonryItem>
            </PageWrapper>
        </div>
    );
}
