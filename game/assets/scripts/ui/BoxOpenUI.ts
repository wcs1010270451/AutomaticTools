import { _decorator, Component, Node, Label, Button, Sprite, tween, Vec3, UIOpacity } from 'cc';
const { ccclass, property } = _decorator;

import { BoxSystem, BoxResult } from '../core/BoxSystem';
import { RARITY_NAMES, RARITY_COLORS, SLOT_NAMES } from '../core/EquipmentSystem';
import { PlayerData } from '../core/PlayerData';
import { UIManager } from './UIManager';

/**
 * BoxOpenUI - 开箱界面
 * 开箱动画 + 结果展示
 */
@ccclass('BoxOpenUI')
export class BoxOpenUI extends Component {
    @property(Node)
    boxNode: Node | null = null;          // 箱子节点（用于动画）

    @property(Node)
    resultPanel: Node | null = null;      // 结果展示面板

    @property(Label)
    resultNameLabel: Label | null = null;

    @property(Label)
    resultRarityLabel: Label | null = null;

    @property(Label)
    resultStatsLabel: Label | null = null;

    @property(Sprite)
    resultIcon: Sprite | null = null;

    @property(Button)
    openBtn: Button | null = null;

    @property(Button)
    backBtn: Button | null = null;

    @property(Label)
    goldLabel: Label | null = null;

    private isAnimating: boolean = false;

    onLoad() {
        this.openBtn?.node.on('click', this.onOpenBox, this);
        this.backBtn?.node.on('click', this.onBack, this);
        if (this.resultPanel) this.resultPanel.active = false;
    }

    public onShow(): void {
        this.refreshGold();
        if (this.resultPanel) this.resultPanel.active = false;
        if (this.boxNode) this.boxNode.active = true;
    }

    private refreshGold(): void {
        const player = PlayerData.getInstance().player;
        if (player && this.goldLabel) {
            this.goldLabel.string = `金币: ${player.gold}`;
        }
    }

    private async onOpenBox(): Promise<void> {
        if (this.isAnimating) return;
        this.isAnimating = true;

        // 播放开箱动画（箱子抖动）
        this.playBoxShake();

        // 请求服务端开箱
        const result = await BoxSystem.getInstance().openBox();

        if (!result) {
            this.isAnimating = false;
            return;
        }

        // 等待抖动动画完成后显示结果
        this.scheduleOnce(() => {
            this.showResult(result);
            this.isAnimating = false;
            this.refreshGold();
            PlayerData.getInstance().refresh();
        }, 0.8);
    }

    private playBoxShake(): void {
        if (!this.boxNode) return;

        tween(this.boxNode)
            .to(0.05, { eulerAngles: new Vec3(0, 0, 5) })
            .to(0.05, { eulerAngles: new Vec3(0, 0, -5) })
            .to(0.05, { eulerAngles: new Vec3(0, 0, 4) })
            .to(0.05, { eulerAngles: new Vec3(0, 0, -4) })
            .to(0.05, { eulerAngles: new Vec3(0, 0, 3) })
            .to(0.05, { eulerAngles: new Vec3(0, 0, -3) })
            .to(0.05, { eulerAngles: new Vec3(0, 0, 0) })
            .start();
    }

    private showResult(result: BoxResult): void {
        if (this.boxNode) this.boxNode.active = false;
        if (this.resultPanel) {
            this.resultPanel.active = true;
            // 结果面板弹出动画
            this.resultPanel.setScale(new Vec3(0.5, 0.5, 1));
            tween(this.resultPanel)
                .to(0.3, { scale: new Vec3(1, 1, 1) }, { easing: 'backOut' })
                .start();
        }

        const equip = result.equipment;
        const rarityName = RARITY_NAMES[equip.rarity] || '未知';
        const slotName = SLOT_NAMES[equip.slot] || equip.slot;

        if (this.resultNameLabel) {
            this.resultNameLabel.string = equip.name;
        }
        if (this.resultRarityLabel) {
            this.resultRarityLabel.string = `[${rarityName}] ${slotName}`;
            // 设置稀有度颜色
            const color = RARITY_COLORS[equip.rarity];
            if (color) {
                this.resultRarityLabel.color.set(color);
            }
        }
        if (this.resultStatsLabel) {
            let stats = `攻击:${equip.atk} 防御:${equip.def}\n生命:${equip.hp} 速度:${equip.spd}`;
            if (equip.bonusAttrs && equip.bonusAttrs.length > 0) {
                const bonusStr = equip.bonusAttrs
                    .map(b => `${b.attr}+${b.value}`)
                    .join(' ');
                stats += `\n附加: ${bonusStr}`;
            }
            this.resultStatsLabel.string = stats;
        }

        // TODO: 根据 equip.iconPath 加载装备图标到 resultIcon
    }

    private onBack(): void {
        UIManager.instance?.closePanel();
    }
}
