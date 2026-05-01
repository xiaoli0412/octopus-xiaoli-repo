type ReadableChannelKeyLike = {
	id?: number | null;
	remark?: string | null;
	channel_key?: string | null;
};

type ChannelKeyLabelOptions = {
	fallbackLabel?: string;
};

export function getChannelKeyLabel(key: ReadableChannelKeyLike, options?: ChannelKeyLabelOptions): string {
	const idPart = typeof key.id === 'number' && key.id > 0 ? ` (#${key.id})` : '';
	const remark = key.remark?.trim();
	if (remark) {
		return `${remark}${idPart}`;
	}
	const rawKey = key.channel_key?.trim() ?? '';
	if (rawKey) {
		const masked = rawKey.length > 10 ? `${rawKey.slice(0, 4)}...${rawKey.slice(-4)}` : rawKey;
		return `${masked}${idPart}`;
	}
	const fallbackLabel = options?.fallbackLabel?.trim() || 'Key';
	return idPart ? `${fallbackLabel}${idPart}` : fallbackLabel;
}

export function buildChannelKeyLabelMap<T extends ReadableChannelKeyLike>(keys: T[], options?: ChannelKeyLabelOptions): Map<number, string> {
	const map = new Map<number, string>();
	for (const key of keys) {
		if (typeof key.id !== 'number' || key.id <= 0) {
			continue;
		}
		map.set(key.id, getChannelKeyLabel(key, options));
	}
	return map;
}
