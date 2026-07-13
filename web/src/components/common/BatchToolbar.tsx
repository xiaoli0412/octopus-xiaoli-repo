'use client';

import { useState } from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { CheckSquare, Square, X } from 'lucide-react';
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { cn } from '@/lib/utils';

export type BatchActionVariant = 'default' | 'destructive';

export type BatchAction = {
	label: string;
	onClick: () => void;
	variant?: BatchActionVariant;
	requireConfirm?: boolean;
	confirmText?: string;
};

export type BatchToolbarProps = {
	selectedIds: number[];
	onClearSelection: () => void;
	actions: BatchAction[];
};

export function BatchToolbar({ selectedIds, onClearSelection, actions }: BatchToolbarProps) {
	const [confirmAction, setConfirmAction] = useState<BatchAction | null>(null);

	const handleActionClick = (action: BatchAction) => {
		if (action.requireConfirm) {
			setConfirmAction(action);
			return;
		}
		action.onClick();
	};

	const handleConfirm = () => {
		if (confirmAction) {
			confirmAction.onClick();
			setConfirmAction(null);
		}
	};

	return (
		<>
			<AnimatePresence>
				{selectedIds.length > 0 && (
					<motion.div
						initial={{ opacity: 0, y: -8 }}
						animate={{ opacity: 1, y: 0 }}
						exit={{ opacity: 0, y: -8 }}
						transition={{ duration: 0.2 }}
						className="flex shrink-0 items-center gap-2 rounded-2xl border border-primary/30 bg-primary/5 p-1.5"
						data-testid="batch-toolbar"
					>
						<div className="flex items-center gap-1.5 px-2 text-xs font-medium text-primary">
							<CheckSquare className="size-4" />
							<span data-testid="batch-selected-count">{selectedIds.length}</span>
						</div>
						<div className="h-5 w-px bg-border/60" />
						<div className="flex flex-1 items-center gap-1.5">
							{actions.map((action) => (
								<button
									key={action.label}
									type="button"
									onClick={() => handleActionClick(action)}
									className={cn(
										'h-8 rounded-xl border px-3 text-xs font-medium transition focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20 focus-visible:ring-offset-2 focus-visible:ring-offset-background',
										action.variant === 'destructive'
											? 'border-destructive/30 bg-destructive/5 text-destructive hover:bg-destructive/10'
											: 'border-border/60 bg-background/60 text-muted-foreground hover:bg-muted/50 hover:text-foreground',
									)}
								>
									{action.label}
								</button>
							))}
						</div>
						<button
							type="button"
							onClick={onClearSelection}
							className="flex h-8 w-8 items-center justify-center rounded-xl border border-border/60 bg-background/60 text-muted-foreground transition hover:bg-muted/50 hover:text-foreground"
							aria-label="clear selection"
						>
							<X className="size-4" />
						</button>
					</motion.div>
				)}
			</AnimatePresence>

			<Dialog open={confirmAction !== null} onOpenChange={(open) => { if (!open) setConfirmAction(null); }}>
				<DialogContent>
					<DialogHeader>
						<DialogTitle>{confirmAction?.label}</DialogTitle>
						<DialogDescription>
							{confirmAction?.confirmText ?? 'Are you sure you want to proceed?'}
						</DialogDescription>
					</DialogHeader>
					<DialogFooter>
						<Button variant="outline" onClick={() => setConfirmAction(null)}>
							Cancel
						</Button>
						<Button
							variant={confirmAction?.variant === 'destructive' ? 'destructive' : 'default'}
							onClick={handleConfirm}
						>
							Confirm
						</Button>
					</DialogFooter>
				</DialogContent>
			</Dialog>
		</>
	);
}

export function BatchCheckbox({
	checked,
	onChange,
}: {
	checked: boolean;
	onChange: (checked: boolean) => void;
}) {
	return (
		<button
			type="button"
			onClick={(e) => {
				e.stopPropagation();
				onChange(!checked);
			}}
			className="flex size-5 shrink-0 items-center justify-center rounded-md border border-border/60 bg-background/80 transition hover:border-primary/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/20"
			aria-label={checked ? 'deselect' : 'select'}
			data-testid="batch-checkbox"
		>
			{checked ? (
				<CheckSquare className="size-4 text-primary" />
			) : (
				<Square className="size-4 text-transparent" />
			)}
		</button>
	);
}
