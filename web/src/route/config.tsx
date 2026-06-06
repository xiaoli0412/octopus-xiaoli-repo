import { lazyWithPreload } from './lazy-with-preload';
import { lazy, ComponentType } from 'react';
import type { LucideIcon } from 'lucide-react';
import { Bot, Home, Radio, Sparkles, FolderTree, Settings, Logs, Activity, Waypoints } from 'lucide-react';

export type LazyComponent = ReturnType<typeof lazy> & {
    preload: () => Promise<{ default: ComponentType<Record<string, never>> }>
};

export interface RouteConfig {
    id: string;
    label: string;
    icon: LucideIcon;
    component: LazyComponent;
}

const Home_Module = lazyWithPreload(() => import('@/components/modules/home').then(m => ({ default: m.Home })));
const Channel_Module = lazyWithPreload(() => import('@/components/modules/channel').then(m => ({ default: m.Channel })));
const Upstream_Module = lazyWithPreload(() => import('@/components/modules/upstream').then(m => ({ default: m.Upstream })));
const Model_Module = lazyWithPreload(() => import('@/components/modules/model').then(m => ({ default: m.Model })));
const Group_Module = lazyWithPreload(() => import('@/components/modules/group').then(m => ({ default: m.Group })));
const Ops_Module = lazyWithPreload(() => import('@/components/modules/ops').then(m => ({ default: m.Ops })));
const AIAutomation_Module = lazyWithPreload(() => import('@/components/modules/ai-automation').then(m => ({ default: m.AIAutomation })));
const Log_Module = lazyWithPreload(() => import('@/components/modules/log').then(m => ({ default: m.Log })));
const Setting_Module = lazyWithPreload(() => import('@/components/modules/setting').then(m => ({ default: m.Setting })));

export const ROUTES: RouteConfig[] = [
    { id: 'home', label: '首页', icon: Home, component: Home_Module },
    { id: 'channel', label: '渠道', icon: Radio, component: Channel_Module },
    { id: 'upstream', label: '上游', icon: Waypoints, component: Upstream_Module },
    { id: 'group', label: '分组', icon: FolderTree, component: Group_Module },
    { id: 'model', label: '价格', icon: Sparkles, component: Model_Module },
    { id: 'ops', label: '运维', icon: Activity, component: Ops_Module },
    { id: 'ai', label: '自动化', icon: Bot, component: AIAutomation_Module },
    { id: 'log', label: '日志', icon: Logs, component: Log_Module },
    { id: 'setting', label: '设置', icon: Settings, component: Setting_Module },
];

export const CONTENT_MAP = ROUTES.reduce((acc, route) => {
    acc[route.id] = route.component;
    return acc;
}, {} as Record<string, LazyComponent>);
