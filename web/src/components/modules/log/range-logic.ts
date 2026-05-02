export function hasPartialTimeRange(startTime: string, endTime: string) {
	return Boolean((startTime && !endTime) || (!startTime && endTime));
}
