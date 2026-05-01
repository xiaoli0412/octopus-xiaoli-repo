const UNKNOWN_VERSION = 'unknown';
const DEVELOPMENT_VERSION = 'dev';
const DEVELOPMENT_VERSION_ALIAS = 'development';

function normalizeVersionForComparison(value?: string | null) {
	const normalized = (value ?? '').trim().toLowerCase();
	if (!normalized) {
		return '';
	}

	const withoutLeadingV = normalized.replace(/^v/, '');
	return withoutLeadingV.replace(/[\s_-]*\((alpha|beta|rc)\)$/i, '-$1');
}

export function formatReleaseTagDisplay(value: string | null | undefined) {
	const normalized = (value ?? '').trim();
	if (!normalized) {
		return '';
	}

	const match = normalized.match(/^v?(\d+(?:\.\d+){1,2})-(alpha|beta|rc)$/i);
	if (match) {
		return `${match[1]}(${match[2].toLowerCase()})`;
	}

	return normalized.replace(/^v/, '');
}

export type VersionInfoLabels = {
	versionUnavailable: string;
	frontendVersionUnknown: string;
	frontendVersionUnknownHint: string;
	developmentVersion: string;
};

export function isUnknownVersion(value?: string | null) {
	return (value ?? '').trim().toLowerCase() === UNKNOWN_VERSION;
}

export function isDevelopmentVersion(value?: string | null) {
	const normalized = (value ?? '').trim().toLowerCase();
	return normalized === DEVELOPMENT_VERSION || normalized === DEVELOPMENT_VERSION_ALIAS;
}

export function formatVersionDisplay(
	value: string | null | undefined,
	fallbackText: string,
	developmentText?: string,
) {
	const normalized = (value ?? '').trim();
	if (!normalized || normalized.toLowerCase() === UNKNOWN_VERSION) {
		return fallbackText;
	}

	if (isDevelopmentVersion(normalized)) {
		return developmentText ?? normalized;
	}

	return normalized;
}

export function getCacheMismatchPresentation(
	frontendVersion: string | null | undefined,
	backendVersion: string | null | undefined,
	labels: VersionInfoLabels,
) {
	return {
		frontendLabel: formatVersionDisplay(frontendVersion, labels.frontendVersionUnknown, labels.developmentVersion),
		backendLabel: formatVersionDisplay(backendVersion, labels.versionUnavailable, labels.developmentVersion),
		note: isUnknownVersion(frontendVersion) ? labels.frontendVersionUnknownHint : '',
	};
}

export function isSameReleaseVersion(left?: string | null, right?: string | null) {
	const leftNormalized = normalizeVersionForComparison(left);
	const rightNormalized = normalizeVersionForComparison(right);
	return !!leftNormalized && !!rightNormalized && leftNormalized === rightNormalized;
}
