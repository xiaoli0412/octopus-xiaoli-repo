'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';
import { User, KeyRound, Lock, Eye, EyeOff, RotateCw } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import { useChangeUsername, useChangePassword, useAuth, useRotateJWTSecret } from '@/api/endpoints/user';
import { toast } from '@/components/common/Toast';

export function SettingAccount() {
    const t = useTranslations('setting');
    const { logout } = useAuth();
    const changeUsername = useChangeUsername();
    const changePassword = useChangePassword();
    const rotateJWTSecret = useRotateJWTSecret();

    const [newUsername, setNewUsername] = useState('');
    const [usernamePassword, setUsernamePassword] = useState('');
    const [oldPassword, setOldPassword] = useState('');
    const [newPassword, setNewPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');

    const [showOldPassword, setShowOldPassword] = useState(false);
    const [showNewPassword, setShowNewPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const [showUsernamePassword, setShowUsernamePassword] = useState(false);

    const handleChangeUsername = () => {
        if (!newUsername.trim()) {
            toast.error(t('account.username.empty'));
            return;
        }
        if (!usernamePassword) {
            toast.error(t('account.password.oldEmpty'));
            return;
        }

        changeUsername.mutate(
            { newUsername: newUsername.trim(), currentPassword: usernamePassword },
            {
                onSuccess: () => {
                    toast.success(t('account.username.success'));
                    setTimeout(() => logout(), 1000);
                },
                onError: () => {
                    toast.error(t('account.username.failed'));
                },
            }
        );
    };

    const handleChangePassword = () => {
        if (!oldPassword) {
            toast.error(t('account.password.oldEmpty'));
            return;
        }
        if (!newPassword) {
            toast.error(t('account.password.newEmpty'));
            return;
        }
        if (newPassword !== confirmPassword) {
            toast.error(t('account.password.mismatch'));
            return;
        }
        if (newPassword.length < 6) {
            toast.error(t('account.password.tooShort'));
            return;
        }

        changePassword.mutate(
            { oldPassword, newPassword },
            {
                onSuccess: () => {
                    toast.success(t('account.password.success'));
                    setTimeout(() => logout(), 1000);
                },
                onError: () => {
                    toast.error(t('account.password.failed'));
                },
            }
        );
    };

    const handleRotateJWTSecret = () => {
        if (!window.confirm(t('account.rotateSecret.confirm'))) {
            return;
        }
        rotateJWTSecret.mutate(undefined, {
            onSuccess: () => {
                toast.success(t('account.rotateSecret.success'));
            },
            onError: () => {
                toast.error(t('account.rotateSecret.failed'));
            },
        });
    };

    return (
        <div className="octo-setting-card">
            <h2 className="octo-setting-heading">
                <User className="size-4" />
                {t('account.title')}
            </h2>

            <div className="grid gap-3 lg:grid-cols-2">
                {/* 修改用户名 */}
                <div className="rounded-2xl border border-border/60 bg-background/55 p-3">
                    <div className="mb-2 flex items-center gap-2 text-sm font-medium text-muted-foreground">
                        <KeyRound className="size-4" />
                        {t('account.username.label')}
                    </div>
                    <div className="space-y-2">
                        <Input
                            value={newUsername}
                            onChange={(e) => setNewUsername(e.target.value)}
                            placeholder={t('account.username.placeholder')}
                            className="h-9 rounded-xl"
                        />
                        <div className="relative">
                            <Input
                                type={showUsernamePassword ? 'text' : 'password'}
                                value={usernamePassword}
                                onChange={(e) => setUsernamePassword(e.target.value)}
                                placeholder={t('account.password.oldPlaceholder')}
                                className="h-9 rounded-xl pr-10"
                            />
                            <button
                                type="button"
                                onClick={() => setShowUsernamePassword(!showUsernamePassword)}
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                            >
                                {showUsernamePassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                            </button>
                        </div>
                        <Button
                            onClick={handleChangeUsername}
                            disabled={changeUsername.isPending || !newUsername.trim() || !usernamePassword}
                            className="h-9 w-full rounded-xl"
                        >
                            {changeUsername.isPending ? t('account.saving') : t('account.save')}
                        </Button>
                    </div>
                </div>

                {/* 修改密码 */}
                <div className="rounded-2xl border border-border/60 bg-background/55 p-3">
                    <div className="mb-2 flex items-center gap-2 text-sm font-medium text-muted-foreground">
                        <Lock className="size-4" />
                        {t('account.password.label')}
                    </div>
                    <div className="space-y-2">
                        <div className="relative">
                            <Input
                                type={showOldPassword ? 'text' : 'password'}
                                value={oldPassword}
                                onChange={(e) => setOldPassword(e.target.value)}
                                placeholder={t('account.password.oldPlaceholder')}
                                className="h-9 rounded-xl pr-10"
                            />
                            <button
                                type="button"
                                onClick={() => setShowOldPassword(!showOldPassword)}
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                            >
                                {showOldPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                            </button>
                        </div>
                        <div className="relative">
                            <Input
                                type={showNewPassword ? 'text' : 'password'}
                                value={newPassword}
                                onChange={(e) => setNewPassword(e.target.value)}
                                placeholder={t('account.password.newPlaceholder')}
                                className="h-9 rounded-xl pr-10"
                            />
                            <button
                                type="button"
                                onClick={() => setShowNewPassword(!showNewPassword)}
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                            >
                                {showNewPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                            </button>
                        </div>
                        <div className="relative">
                            <Input
                                type={showConfirmPassword ? 'text' : 'password'}
                                value={confirmPassword}
                                onChange={(e) => setConfirmPassword(e.target.value)}
                                placeholder={t('account.password.confirmPlaceholder')}
                                className="h-9 rounded-xl pr-10"
                            />
                            <button
                                type="button"
                                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground transition-colors hover:text-foreground"
                            >
                                {showConfirmPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
                            </button>
                        </div>
                        <Button
                            onClick={handleChangePassword}
                            disabled={changePassword.isPending || !oldPassword || !newPassword || !confirmPassword}
                            className="h-9 w-full rounded-xl"
                        >
                            {changePassword.isPending ? t('account.saving') : t('account.password.change')}
                        </Button>
                    </div>
                </div>
            </div>

            {/* 轮换 JWT 密钥 */}
            <div className="rounded-2xl border border-border/60 bg-background/55 p-3">
                <div className="mb-2 flex items-center gap-2 text-sm font-medium text-muted-foreground">
                    <RotateCw className="size-4" />
                    {t('account.rotateSecret.label')}
                </div>
                <p className="mb-2 text-xs text-muted-foreground">
                    {t('account.rotateSecret.hint')}
                </p>
                <Button
                    onClick={handleRotateJWTSecret}
                    disabled={rotateJWTSecret.isPending}
                    variant="outline"
                    className="h-9 w-full rounded-xl"
                >
                    {rotateJWTSecret.isPending ? t('account.saving') : t('account.rotateSecret.button')}
                </Button>
            </div>
        </div>
    );
}

