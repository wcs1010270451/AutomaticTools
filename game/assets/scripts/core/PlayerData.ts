/**
 * PlayerData - 玩家数据管理
 * 管理玩家角色信息、属性、经验等级
 */

import { ApiClient } from '../network/ApiClient';

export interface PlayerInfo {
    id: number;
    userId: number;
    nickname: string;
    level: number;
    exp: number;
    gold: number;
    combatPower: number;
    wins: number;
    losses: number;
}

export class PlayerData {
    private static instance: PlayerData;
    private api: ApiClient;
    private _player: PlayerInfo | null = null;

    public static getInstance(): PlayerData {
        if (!PlayerData.instance) {
            PlayerData.instance = new PlayerData();
        }
        return PlayerData.instance;
    }

    constructor() {
        this.api = ApiClient.getInstance();
    }

    get player(): PlayerInfo | null {
        return this._player;
    }

    /**
     * 初始化/获取玩家数据
     */
    public async init(): Promise<PlayerInfo | null> {
        const result = await this.api.post<PlayerInfo>('/game/init');
        if (result.data) {
            this._player = result.data;
            return this._player;
        }
        return null;
    }

    /**
     * 获取当前经验升级所需值
     */
    public getExpToNextLevel(): number {
        if (!this._player) return 0;
        return this._player.level * 100;
    }

    /**
     * 获取经验进度百分比 (0-1)
     */
    public getExpProgress(): number {
        if (!this._player) return 0;
        const needed = this.getExpToNextLevel();
        return needed > 0 ? this._player.exp / needed : 0;
    }

    /**
     * 刷新玩家数据
     */
    public async refresh(): Promise<void> {
        await this.init();
    }
}
