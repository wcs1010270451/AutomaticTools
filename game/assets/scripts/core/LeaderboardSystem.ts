/**
 * LeaderboardSystem - 排行榜系统
 * 获取战力榜和胜场榜数据
 */

import { ApiClient } from '../network/ApiClient';

export interface LeaderboardEntry {
    rank: number;
    playerId: number;
    nickname: string;
    level: number;
    score: number;
    combatPower?: number;
    wins?: number;
}

export type RankType = 'power' | 'wins';

export class LeaderboardSystem {
    private static instance: LeaderboardSystem;
    private api: ApiClient;

    public static getInstance(): LeaderboardSystem {
        if (!LeaderboardSystem.instance) {
            LeaderboardSystem.instance = new LeaderboardSystem();
        }
        return LeaderboardSystem.instance;
    }

    constructor() {
        this.api = ApiClient.getInstance();
    }

    /**
     * 获取排行榜数据
     * @param type 排行类型: 'power' 战力榜, 'wins' 胜场榜
     * @param limit 获取条数，默认50
     */
    public async getLeaderboard(type: RankType = 'power', limit: number = 50): Promise<LeaderboardEntry[]> {
        const result = await this.api.get<LeaderboardEntry[]>('/game/leaderboard', {
            type,
            limit: String(limit),
        });
        return result.data || [];
    }
}
