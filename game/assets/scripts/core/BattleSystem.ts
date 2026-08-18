/**
 * BattleSystem - 战斗系统
 * 处理异步PK：对手匹配、战斗请求、战斗回放数据
 */

import { ApiClient } from '../network/ApiClient';

export interface Opponent {
    playerId: number;
    nickname: string;
    level: number;
    combatPower: number;
    wins: number;
}

export interface BattleRound {
    round: number;
    attackerDmg: number;
    defenderDmg: number;
    attackerHp: number;
    defenderHp: number;
}

export interface BattleResult {
    winnerId: number;
    battleLog: BattleRound[];
    rewardGold: number;
    rewardExp: number;
    isWin: boolean;
}

export class BattleSystem {
    private static instance: BattleSystem;
    private api: ApiClient;
    private _opponents: Opponent[] = [];

    public static getInstance(): BattleSystem {
        if (!BattleSystem.instance) {
            BattleSystem.instance = new BattleSystem();
        }
        return BattleSystem.instance;
    }

    constructor() {
        this.api = ApiClient.getInstance();
    }

    get opponents(): Opponent[] {
        return this._opponents;
    }

    /**
     * 获取可挑战的对手列表
     */
    public async getOpponents(): Promise<Opponent[]> {
        const result = await this.api.get<Opponent[]>('/game/opponents');
        if (result.data) {
            this._opponents = result.data;
        }
        return this._opponents;
    }

    /**
     * 发起挑战
     * @param defenderId 对手的玩家ID
     */
    public async challenge(defenderId: number): Promise<BattleResult | null> {
        const result = await this.api.post<BattleResult>('/game/challenge', { defenderId });
        if (result.data) {
            return result.data;
        }
        return null;
    }
}
