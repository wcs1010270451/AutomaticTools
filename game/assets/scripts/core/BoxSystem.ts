/**
 * BoxSystem - 开箱系统
 * 处理开箱请求和结果
 */

import { ApiClient } from '../network/ApiClient';
import { EquipmentItem } from './EquipmentSystem';

export interface BoxResult {
    equipment: EquipmentItem;
    isPity: boolean;
}

export class BoxSystem {
    private static instance: BoxSystem;
    private api: ApiClient;

    public static getInstance(): BoxSystem {
        if (!BoxSystem.instance) {
            BoxSystem.instance = new BoxSystem();
        }
        return BoxSystem.instance;
    }

    constructor() {
        this.api = ApiClient.getInstance();
    }

    /**
     * 开箱
     * @returns 开箱结果，包含获得的装备和是否触发保底
     */
    public async openBox(): Promise<BoxResult | null> {
        const result = await this.api.post<BoxResult>('/game/box/open');
        if (result.data) {
            return result.data;
        }
        return null;
    }

    /**
     * 批量开箱（连续开多个）
     * @param count 开箱次数
     */
    public async openBoxMultiple(count: number): Promise<BoxResult[]> {
        const results: BoxResult[] = [];
        for (let i = 0; i < count; i++) {
            const result = await this.openBox();
            if (result) {
                results.push(result);
            } else {
                break;
            }
        }
        return results;
    }
}
