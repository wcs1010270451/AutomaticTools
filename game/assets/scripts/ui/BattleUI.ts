import { _decorator, Component, Node, Label, Button, ScrollView, Prefab, instantiate, tween, Vec3 } from 'cc';
const { ccclass, property } = _decorator;

import { BattleSystem, Opponent, BattleResult } from '../core/BattleSystem';
import { PlayerData } from '../core/PlayerData';
import { UIManager } from './UIManager';

/**
 * BattleUI - PK 对战界面
 * 对手列表 + 战斗回放 + 结算
 */
@ccclass('BattleUI')
export class BattleUI extends Component {
    @property(ScrollView)
    opponentScroll: ScrollView | null = null;

    @property(Prefab)
    opponentItemPrefab: Prefab | null = null;

    @property(Node)
    battlePanel: Node | null = null;      // 战斗回放区域

    @property(Node)
    resultPanel: Node | null = null;      // 结算面板

    @property(Label)
    battleLogLabel: Label | null = null;

    @property(Label)
    resultLabel: Label | null = null;

    @property(Label)
    rewardLabel: Label | null = null;

    @property(Button)
    backBtn: Button | null = null;

    @property(Button)
    confirmBtn: Button | null = null;

    onLoad() {
        this.backBtn?.node.on('click', this.onBack, this);
        this.confirmBtn?.node.on('click', this.onConfirm, this);
        if (this.battlePanel) this.battlePanel.active = false;
        if (this.resultPanel) this.resultPanel.active = false;
    }

    public async onShow(): Promise<void> {
        await this.loadOpponents();
        if (this.battlePanel) this.battlePanel.active = false;
        if (this.resultPanel) this.resultPanel.active = false;
    }

    private async loadOpponents(): Promise<void> {
        const system = BattleSystem.getInstance();
        const opponents = await system.getOpponents();

        const content = this.opponentScroll?.content;
        if (!content) return;
        content.removeAllChildren();

        for (const opp of opponents) {
            this.createOpponentItem(content, opp);
        }
    }

    private createOpponentItem(parent: Node, opp: Opponent): void {
        let itemNode: Node;
        if (this.opponentItemPrefab) {
            itemNode = instantiate(this.opponentItemPrefab);
        } else {
            itemNode = new Node('OpponentItem');
        }
        parent.addChild(itemNode);

        const label = itemNode.getComponentInChildren(Label);
        if (label) {
            label.string = `${opp.nickname} | Lv.${opp.level} | 战力:${opp.combatPower}`;
        }

        // 点击挑战
        itemNode.on('click', () => {
            this.startBattle(opp);
        });
    }

    private async startBattle(opponent: Opponent): Promise<void> {
        // 显示战斗面板
        if (this.battlePanel) this.battlePanel.active = true;
        if (this.battleLogLabel) this.battleLogLabel.string = '战斗中...';

        const result = await BattleSystem.getInstance().challenge(opponent.playerId);

        if (!result) {
            if (this.battleLogLabel) this.battleLogLabel.string = '挑战失败，请稍后重试';
            return;
        }

        // 播放战斗回放
        this.playBattleReplay(result);
    }

    private playBattleReplay(result: BattleResult): void {
        if (!this.battleLogLabel) return;

        const log = result.battleLog;
        let roundIndex = 0;

        const showNextRound = () => {
            if (roundIndex >= log.length) {
                // 回放结束，显示结算
                this.scheduleOnce(() => this.showResult(result), 0.5);
                return;
            }

            const round = log[roundIndex];
            this.battleLogLabel!.string =
                `回合 ${round.round}\n` +
                `你造成 ${round.attackerDmg} 伤害\n` +
                `对手造成 ${round.defenderDmg} 伤害\n` +
                `你的HP: ${round.attackerHp} | 对手HP: ${round.defenderHp}`;
            roundIndex++;

            this.scheduleOnce(showNextRound, 0.3);
        };

        showNextRound();
    }

    private showResult(result: BattleResult): void {
        if (this.battlePanel) this.battlePanel.active = false;
        if (this.resultPanel) this.resultPanel.active = true;

        if (this.resultLabel) {
            this.resultLabel.string = result.isWin ? '胜利！' : '失败...';
        }
        if (this.rewardLabel) {
            this.rewardLabel.string = `金币 +${result.rewardGold}  经验 +${result.rewardExp}`;
        }

        // 刷新玩家数据
        PlayerData.getInstance().refresh();
    }

    private onConfirm(): void {
        if (this.resultPanel) this.resultPanel.active = false;
        this.loadOpponents();
    }

    private onBack(): void {
        UIManager.instance?.closePanel();
    }
}
