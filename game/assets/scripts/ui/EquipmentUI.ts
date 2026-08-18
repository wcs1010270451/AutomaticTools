import { _decorator, Component, Node, Label, Button, ScrollView, Prefab, instantiate, Sprite } from 'cc';
const { ccclass, property } = _decorator;

import { EquipmentSystem, EquipmentItem, RARITY_NAMES, RARITY_COLORS, SLOT_NAMES } from '../core/EquipmentSystem';
import { UIManager } from './UIManager';

/**
 * EquipmentUI - 装备背包界面
 * 展示装备列表、穿戴/替换功能
 */
@ccclass('EquipmentUI')
export class EquipmentUI extends Component {
    @property(ScrollView)
    scrollView: ScrollView | null = null;

    @property(Prefab)
    equipmentItemPrefab: Prefab | null = null;

    @property(Node)
    equippedPanel: Node | null = null;    // 已装备栏位

    @property(Button)
    backBtn: Button | null = null;

    // 已装备槽位显示
    @property(Label)
    weaponLabel: Label | null = null;

    @property(Label)
    helmetLabel: Label | null = null;

    @property(Label)
    armorLabel: Label | null = null;

    @property(Label)
    shieldLabel: Label | null = null;

    @property(Label)
    accessoryLabel: Label | null = null;

    onLoad() {
        this.backBtn?.node.on('click', this.onBack, this);
    }

    public async onShow(): Promise<void> {
        await this.loadEquipments();
    }

    private async loadEquipments(): Promise<void> {
        const system = EquipmentSystem.getInstance();
        const equipments = await system.loadEquipments();

        // 更新已装备栏位
        this.updateEquippedDisplay();

        // 清空列表并填充
        const content = this.scrollView?.content;
        if (!content) return;
        content.removeAllChildren();

        for (const item of equipments) {
            this.createEquipmentItem(content, item);
        }
    }

    private updateEquippedDisplay(): void {
        const equipped = EquipmentSystem.getInstance().equippedItems;
        const labels: Record<string, Label | null> = {
            weapon: this.weaponLabel,
            helmet: this.helmetLabel,
            armor: this.armorLabel,
            shield: this.shieldLabel,
            accessory: this.accessoryLabel,
        };

        for (const [slot, label] of Object.entries(labels)) {
            if (label) {
                const item = equipped[slot];
                label.string = item ? `${item.name}` : `[空] ${SLOT_NAMES[slot]}`;
            }
        }
    }

    private createEquipmentItem(parent: Node, item: EquipmentItem): void {
        let itemNode: Node;
        if (this.equipmentItemPrefab) {
            itemNode = instantiate(this.equipmentItemPrefab);
        } else {
            itemNode = new Node('EquipmentItem');
        }
        parent.addChild(itemNode);

        // 设置显示内容
        const label = itemNode.getComponentInChildren(Label);
        if (label) {
            const rarityName = RARITY_NAMES[item.rarity] || '';
            const equippedTag = item.equipped ? ' [已装备]' : '';
            label.string = `[${rarityName}] ${item.name}${equippedTag}`;
        }

        // 点击穿戴
        itemNode.on('click', async () => {
            if (!item.equipped) {
                await EquipmentSystem.getInstance().equip(item.id);
                await this.loadEquipments(); // 刷新列表
            }
        });
    }

    private onBack(): void {
        UIManager.instance?.closePanel();
    }
}
