/**
 * ApiClient - HTTP 请求封装
 * 负责与后端 API 通信，支持 Token 认证
 */

export interface ApiResponse<T = any> {
    data?: T;
    error?: string;
    requestId?: string;
}

export class ApiClient {
    private static instance: ApiClient;
    private baseUrl: string = '';
    private token: string = '';

    public static getInstance(): ApiClient {
        if (!ApiClient.instance) {
            ApiClient.instance = new ApiClient();
        }
        return ApiClient.instance;
    }

    /**
     * 初始化 API 客户端
     * @param baseUrl 后端 API 地址
     */
    public init(baseUrl: string): void {
        this.baseUrl = baseUrl.replace(/\/$/, '');
    }

    /**
     * 设置认证 Token
     */
    public setToken(token: string): void {
        this.token = token;
    }

    public getToken(): string {
        return this.token;
    }

    /**
     * GET 请求
     */
    public async get<T>(path: string, params?: Record<string, string>): Promise<ApiResponse<T>> {
        let url = `${this.baseUrl}/api${path}`;
        if (params) {
            const query = Object.entries(params)
                .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
                .join('&');
            url += `?${query}`;
        }
        return this.request<T>(url, 'GET');
    }

    /**
     * POST 请求
     */
    public async post<T>(path: string, body?: any): Promise<ApiResponse<T>> {
        const url = `${this.baseUrl}/api${path}`;
        return this.request<T>(url, 'POST', body);
    }

    private async request<T>(url: string, method: string, body?: any): Promise<ApiResponse<T>> {
        return new Promise((resolve, reject) => {
            const xhr = new XMLHttpRequest();
            xhr.open(method, url, true);
            xhr.setRequestHeader('Content-Type', 'application/json');
            if (this.token) {
                xhr.setRequestHeader('Authorization', `Bearer ${this.token}`);
            }

            xhr.onreadystatechange = () => {
                if (xhr.readyState !== 4) return;

                if (xhr.status >= 200 && xhr.status < 300) {
                    try {
                        const data = JSON.parse(xhr.responseText);
                        resolve(data as ApiResponse<T>);
                    } catch (e) {
                        reject(new Error('响应解析失败'));
                    }
                } else {
                    try {
                        const err = JSON.parse(xhr.responseText);
                        resolve({ error: err.error || '请求失败' });
                    } catch {
                        resolve({ error: `HTTP ${xhr.status}` });
                    }
                }
            };

            xhr.onerror = () => reject(new Error('网络连接失败'));
            xhr.send(body ? JSON.stringify(body) : null);
        });
    }
}
