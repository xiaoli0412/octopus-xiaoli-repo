import type { BillingMode, ProbePolicy } from '@/api/endpoints/model';

export type ChannelKeySourceTypeKey = 'unknown' | 'publicFree' | 'paidMetered' | 'privateInternal';

export function getChannelKeySourceTypeKey(value?: string | null): ChannelKeySourceTypeKey {
    switch ((value ?? '').trim().toLowerCase()) {
        case 'public/free':
            return 'publicFree';
        case 'paid/metered':
            return 'paidMetered';
        case 'private/internal':
            return 'privateInternal';
        default:
            return 'unknown';
    }
}

export function getBillingModeKey(value?: string | null): BillingMode {
    switch ((value ?? '').trim().toLowerCase()) {
        case 'per_request':
            return 'per_request';
        case 'per_token':
            return 'per_token';
        case 'per_quota':
            return 'per_quota';
        case 'flat':
            return 'flat';
        case 'free':
            return 'free';
        default:
            return 'unknown';
    }
}

export function getProbePolicyKey(value?: string | null): ProbePolicy {
    switch ((value ?? '').trim().toLowerCase()) {
        case 'sparse_single':
            return 'sparse_single';
        case 'sequential':
            return 'sequential';
        case 'concurrent':
            return 'concurrent';
        default:
            return 'passive_only';
    }
}
