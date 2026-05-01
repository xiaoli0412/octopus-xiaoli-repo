export const AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY = 'octopus-ai-automation-focus-target';

export type AIAutomationFocusTarget = 'learning';

function canUseSessionStorage() {
    return typeof window !== 'undefined' && typeof window.sessionStorage !== 'undefined';
}

export function queueAIAutomationFocusTarget(target: AIAutomationFocusTarget) {
    if (!canUseSessionStorage()) return;
    window.sessionStorage.setItem(AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY, target);
}

export function consumeAIAutomationFocusTarget(): AIAutomationFocusTarget | null {
    if (!canUseSessionStorage()) return null;

    const value = window.sessionStorage.getItem(AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY);
    window.sessionStorage.removeItem(AI_AUTOMATION_FOCUS_TARGET_STORAGE_KEY);

    if (value === 'learning') return value;
    return null;
}
