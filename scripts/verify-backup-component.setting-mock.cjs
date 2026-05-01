const React = require('../web/node_modules/react');

function getState() {
	return global.__backupComponentVerifyState;
}

module.exports = {
	useExportDB() {
		const state = getState();
		return {
			async mutateAsync(payload = {}) {
				state.exportMutateAsyncCalls.push(payload);
				return { filename: 'ignored.json' };
			},
			isPending: false,
		};
	},
	useImportDB() {
		const state = getState();
		return {
			data: state.importDBState.data,
			isPending: false,
			reset() {
				state.importDBState.data = undefined;
			},
			async mutateAsync(payload) {
				state.importMutateAsyncCalls.push(payload);
				state.importCallCount += 1;

				if (payload.dryRun) {
					const legacyModelTarget = payload.modelMappings?.['legacy-model'] ?? 'gpt-4o';
					const mapPreviewToken = payload.file?.name === 'snapshot-map-reset.json' && legacyModelTarget === 'gpt-4.1-mini'
						? 'preview-token-map-updated'
						: 'preview-token-map';
					const dryRunResult = payload.file?.name === 'snapshot-missing-preview-token.json'
						? {
							rows_affected: { channels: 1 },
							dry_run: true,
							mode: payload.mode,
							compatibility: { conflicts: ['channel conflict'] },
						}
						: payload.mode === 'map'
						? {
							rows_affected: { channels: 1, groups: 1 },
							preview_token: mapPreviewToken,
							dry_run: true,
							mode: payload.mode,
							compatibility: {
								alias_preview_mappings: [
									{
										snapshot_model: 'legacy-vision',
										current_model: 'gpt-4.1',
										canonical: 'gpt-4.1',
										contexts: ['routing'],
									},
								],
								missing_providers: ['legacy-provider'],
								model_mapping_previews: [
									{
										source_model: 'legacy-model',
										target_model: legacyModelTarget,
										contexts: ['routing'],
										touched_fields: ['primary_model'],
										usage_count: 2,
										used: true,
										target_exists: true,
									},
									{
										source_model: 'missing-model',
										target_model: 'gpt-4.1-mini',
										contexts: ['fallback'],
										touched_fields: ['fallback_model'],
										usage_count: 1,
										used: true,
										target_exists: false,
										warnings: ['current model not found'],
									},
									{
										source_model: 'unused-model',
										target_model: 'gpt-4.1',
										contexts: ['api_keys'],
										touched_fields: ['model'],
										usage_count: 0,
										used: false,
										target_exists: true,
									},
								],
								credential_rebind_targets: [
									{
										target_type: 'channel_key',
										channel_name: 'Primary',
										key_name: 'key-1',
										models: ['legacy-model'],
										affected_groups: ['group-a'],
									},
								],
							},
						}
						: payload.mode === 'replace'
							? {
								rows_affected: { channels: 1, groups: 1, api_keys: 1 },
								preview_token: 'preview-token-replace',
								dry_run: true,
								mode: payload.mode,
								replace_prune_preview: {
									pruned_channels: ['legacy-channel'],
									pruned_api_keys: ['client-key'],
									warnings: ['Current client keys remain unless the snapshot carries plaintext credentials.'],
								},
								compatibility: {
									conflicts: ['replace conflict'],
									credential_rebind_targets: [
										{
											target_type: 'channel_key',
											channel_name: 'Primary',
											key_name: 'key-1',
											models: ['gpt-4o'],
											affected_groups: ['group-a'],
										},
									],
									replace_pruned_channels: ['legacy-channel'],
									replace_pruned_api_keys: ['client-key'],
								},
							}
						: {
							rows_affected: { channels: 1 },
							preview_token: 'preview-token-1',
							dry_run: true,
							mode: payload.mode,
							compatibility: { conflicts: ['channel conflict'] },
						};
					state.importDBState.data = dryRunResult;
					return dryRunResult;
				}

				const applyResult = {
					rows_affected: { channels: 1, groups: 1 },
					dry_run: false,
					mode: payload.mode,
					post_import_validation: {
						degraded_groups: ['group-a'],
						health_check: {
							summary: { targets: 3, passed: 2, failed: 1, skipped: 0, rate_limited: 0 },
							checks: [
								{ channel_id: 1, channel_name: 'Primary', model: 'gpt-4o', passed: true },
								{ channel_id: 2, channel_name: 'Fallback', model: 'gpt-4.1-mini', passed: true },
								{ channel_id: 3, channel_name: 'Legacy', model: 'gpt-4.1', passed: false },
							],
						},
					},
				};
				state.importDBState.data = applyResult;
				return applyResult;
			},
		};
	},
	useImportSnapshots() {
		const state = getState();
		return {
			data: state.importSnapshotsState.data,
			isLoading: state.importSnapshotsState.isLoading,
			isError: state.importSnapshotsState.isError,
			isFetching: state.importSnapshotsState.isFetching,
			refetch: async () => {
				state.importSnapshotsRefetchCalls?.push({ source: 'verify-backup-component' });
				return { data: state.importSnapshotsState.data };
			},
		};
	},
	usePreviewRollbackImportSnapshot() {
		const state = getState();
		const [data, setData] = React.useState(state.previewRollbackState.data);
		const [isPending, setIsPending] = React.useState(false);
		return {
			data,
			isPending,
			reset() {
				state.previewRollbackState.data = undefined;
				setData(undefined);
			},
			async mutateAsync(payload) {
				const { snapshotName, importScopes } = payload;
				state.previewRollbackCalls.push({ snapshotName, importScopes });
				setIsPending(true);
				const previewResult = {
					snapshot_name: snapshotName,
					applied_scopes: importScopes,
					manifest: { contains_secrets: true, schema_version: '10' },
					rows_summary: { channels: 2, groups: 1 },
					preview_warnings: ['provider mismatch'],
					compatibility: {
						conflicts: ['channel conflict'],
						credential_rebind_targets: [{
							target_type: 'channel_key',
							channel_name: 'Primary',
							key_name: 'key-1',
							models: ['gpt-4o'],
							affected_groups: ['group-a'],
						}],
						missing_providers: ['provider-x'],
						affected_groups: ['group-a'],
						affected_channels: ['channel-a'],
					},
				};
				if (state.previewRollbackDeferred) {
					const deferred = state.previewRollbackDeferred;
					state.previewRollbackDeferred = null;
					await deferred.promise;
				}
				state.previewRollbackState.data = previewResult;
				setData(previewResult);
				setIsPending(false);
				return previewResult;
			},
		};
	},
	useRollbackLatestImportSnapshot() {
		const state = getState();
		const [isPending, setIsPending] = React.useState(false);
		return {
			isPending,
			mutateAsync: async (payload) => {
				state.rollbackLatestCalls.push(payload);
				setIsPending(true);
				try {
					if (state.rollbackLatestDeferred) {
						const deferred = state.rollbackLatestDeferred;
						state.rollbackLatestDeferred = null;
						await deferred.promise;
					}
					return { snapshot_name: 'snapshot-latest' };
				} finally {
					setIsPending(false);
				}
			},
		};
	},
	useRollbackImportSnapshot() {
		const state = getState();
		const [isPending, setIsPending] = React.useState(false);
		return {
			isPending,
			mutateAsync: async (payload) => {
				state.rollbackImportCalls.push(payload);
				setIsPending(true);
				try {
					if (state.rollbackImportDeferred) {
						const deferred = state.rollbackImportDeferred;
						state.rollbackImportDeferred = null;
						await deferred.promise;
					}
					return {
						snapshot_name: payload.snapshotName,
						applied_scopes: payload.importScopes,
					};
				} finally {
					setIsPending(false);
				}
			},
		};
	},
};
