/**
 * EquipmentSystem - 装备系统
 * 管理装备列表、穿戴、属性展示
 */

import { ApiClient } from '../network/ApiClient';

export interface BonusAttr {
    attr: string;
    value: number;
}

export interface EquipmentItem {
    id: number;
    templateId: number;
    name: string;
    slot: string;       // weapon, helmet, armor, shield, accessory
    rarity: number;     // 1-5
    atk: number;
    def: number;
    hp: number;
    spd: number;
    bonusAttrs: BonusAttr[];
    equipped: boolean;
    iconPath: string;
}

/** 稀有度名称映射 */
export const RARITY_NAMES: Record<number, string> = {
    1: '普通',
    2: '优秀',
    3: '稀有',
    4: '史诗',
    5: '传说',
};

/** 稀有度颜色映射 */
export const RARITY_COLORS: Record<number, string> = {
    1: '#FFFFFF',   // 白
    2: '#4CAF50',   // 绿
    3: '#2196F3',   // 蓝
    4: '#9C27B0',   // 紫
    5: '#FF9800',   // 橙
};

/** 装备槽位名称 */
export const SLOT_NAMES: Record<string, string> = {
    weapon: '武器',
    helmet: '头盔',
    armor: '铠甲',
    shield: '盾牌',
    accessory: '饰品',
};

export class EquipmentSystem {
    private static instance: EquipmentSystem;
    private api: ApiClient;
    private _equipments: EquipmentItem[] = [];

    public static getInstance(): EquipmentSystem {
        if (!EquipmentSystem.instance) {
            EquipmentSystem.instance = new EquipmentSystem();
        }
        return EquipmentSystem.instance;
    }

    constructor() {
        this.api = ApiClient.getInstance();
    }

    get equipments(): EquipmentItem[] {
        return this._equipments;
    }

    /**
     * 获取当前已装备的物品（按槽位）
     */
    get equippedItems(): Record<string, EquipmentItem | null> {
        const result: Record<string, EquipmentItem | null> = {
            weapon: null,
            helmet: null,
            armor: null,
            shield: null,
            accessory: null,
        };
        for (const item of this._equipments) {
            if (item.equipped) {
                result[item.slot] = item;
            }
        }
        return result;
    }

    /**
     * 加载装备列表
     */
    public async loadEquipments(): Promise<EquipmentItem[]> {
        const result = await this.api.get<EquipmentItem[]>('/game/equipments');
        if (result.data) {
            this._equipments = result.data;
        }
        return this._equipments;
    }

    /**
     * 穿戴装备
     */
    public async equip(equipmentId: number): Promise<boolean> {
        const result = await this.api.post('/game/equipments/equip', { equipmentId });
        if (!result.error) {
            await this.loadEquipments();
            return true;
        }
        return false;
    }

    /**
     * 按槽位筛选装备
     */
    public getBySlot(slot: string): EquipmentItem[] {
        return this._equipments.filter(e => e.slot === slot);
    }

    /**
     * 按稀有度筛选装备
     */
    public getByRarity(rarity: number): EquipmentItem[] {
        return this._equipments.filter(e => e.rarity === rarity);
    }
}
