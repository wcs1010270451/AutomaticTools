/**
 * AuthService - 登录认证服务
 * 支持微信小游戏 OAuth 和 H5 邮箱/手机号登录
 */

import { ApiClient } from './ApiClient';

export interface LoginResult {
    token: string;
    userId: number;
}

export class AuthService {
    private static instance: AuthService;
    private api: ApiClient;

    public static getInstance(): AuthService {
        if (!AuthService.instance) {
            AuthService.instance = new AuthService();
        }
        return AuthService.instance;
    }

    constructor() {
        this.api = ApiClient.getInstance();
    }

    /**
     * 微信小游戏登录
     * 调用 wx.login 获取 code，发送到后端换取 token
     */
    public async loginWithWechat(): Promise<LoginResult | null> {
        // @ts-ignore - wx 是微信小游戏全局对象
        if (typeof wx === 'undefined') {
            console.warn('非微信环境，无法使用微信登录');
            return null;
        }

        return new Promise((resolve) => {
            // @ts-ignore
            wx.login({
                success: async (res: any) => {
                    if (res.code) {
                        const result = await this.api.post<LoginResult>('/game/auth/wechat', {
                            code: res.code
                        });
                        if (result.data) {
                            this.api.setToken(result.data.token);
                            resolve(result.data);
                        } else {
                            resolve(null);
                        }
                    } else {
                        resolve(null);
                    }
                },
                fail: () => resolve(null)
            });
        });
    }

    /**
     * 邮箱验证码登录 (H5)
     */
    public async loginWithEmail(email: string, code: string): Promise<LoginResult | null> {
        const result = await this.api.post<LoginResult>('/auth/login', {
            email,
            code
        });
        if (result.data) {
            this.api.setToken(result.data.token);
            return result.data;
        }
        return null;
    }

    /**
     * 发送邮箱验证码
     */
    public async sendEmailCode(email: string): Promise<boolean> {
        const result = await this.api.post('/auth/email-code', { email });
        return !result.error;
    }

    /**
     * 从本地存储恢复 Token
     */
    public restoreSession(): boolean {
        const token = localStorage.getItem('game_token');
        if (token) {
            this.api.setToken(token);
            return true;
        }
        return false;
    }

    /**
     * 保存 Token 到本地
     */
    public saveToken(token: string): void {
        localStorage.setItem('game_token', token);
    }

    /**
     * 退出登录
     */
    public logout(): void {
        localStorage.removeItem('game_token');
        this.api.setToken('');
    }
}
