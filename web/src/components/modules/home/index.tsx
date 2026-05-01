'use client';

import { Activity } from './activity';
import { Total } from './total';
import { StatsChart } from './chart';
import { Rank } from './rank';
import { TokenBreakdown } from './token-breakdown';
import { PageWrapper } from '@/components/common/PageWrapper';

export function Home() {
    return (
        <PageWrapper data-testid="home-page" className="h-full min-h-0 overflow-y-auto overscroll-contain space-y-5 pb-24 md:pb-4 rounded-t-3xl md:space-y-6">
            <Total />
            <div data-testid="home-main-grid" className="space-y-5 md:space-y-6">
                <Activity />
                <StatsChart />
                <Rank />
                <TokenBreakdown />
            </div>
        </PageWrapper>
    );
}
