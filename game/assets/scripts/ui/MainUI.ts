import { _decorator, Component, Node, Label, Button, Sprite, ProgressBar } from 'cc';
const { ccclass, property } = _decorator;

import { PlayerData } from '../core/PlayerData';
import { UIManager, UI_PANEL } from './UIManager';

/**
 * MainUI - 游戏主界面
 * 显示玩家信息、导航按钮
 */
@ccclass('MainUI')
export class MainUI extends Component {
    @property(Label)
    nameLabel: Label | null = null;

    @property(Label)
    levelLabel: Label | null = null;

    @property(Label)
    goldLabel: Label | null = null;

    @property(Label)
    powerLabel: Label | null = null;

    @property(Label)
    winLabel: Label | null = null;

    @property(ProgressBar)
    expBar: ProgressBar | null = null;

    @property(Button)
    boxBtn: Button | null = null;

    @property(Button)
    equipBtn: Button | null = null;

    @property(Button)
    battleBtn: Button | null = null;

    @property(Button)
    rankBtn: Button | null = null;

    onLoad() {
        this.boxBtn?.node.on('click', this.onBoxClick, this);
        this.equipBtn?.node.on('click', this.onEquipClick, this);
        this.battleBtn?.node.on('click', this.onBattleClick, this);
        this.rankBtn?.node.on('click', this.onRankClick, this);
    }

    /** 面板显示时刷新数据 */
    public onShow(): void {
        this.refreshUI();
    }

    private refreshUI(): void {
        const player = PlayerData.getInstance().player;
        if (!player) return;

        if (this.nameLabel) this.nameLabel.string = player.nickname;
        if (this.levelLabel) this.levelLabel.string = `Lv.${player.level}`;
        if (this.goldLabel) this.goldLabel.string = `${player.gold}`;
        if (this.powerLabel) this.powerLabel.string = `战力: ${player.combatPower}`;
        if (this.winLabel) this.winLabel.string = `${player.wins}胜 / ${player.losses}负`;
        if (this.expBar) {
            this.expBar.progress = PlayerData.getInstance().getExpProgress();
        }
    }

    private onBoxClick(): void {
        UIManager.instance?.openPanel(UI_PANEL.BOX_OPEN);
    }

    private onEquipClick(): void {
        UIManager.instance?.openPanel(UI_PANEL.EQUIPMENT);
    }

    private onBattleClick(): void {
        UIManager.instance?.openPanel(UI_PANEL.BATTLE);
    }

    private onRankClick(): void {
        UIManager.instance?.openPanel(UI_PANEL.LEADERBOARD);
    }
}
