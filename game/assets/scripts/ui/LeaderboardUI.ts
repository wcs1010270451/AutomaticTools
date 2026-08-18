import { _decorator, Component, Node, Label, Button, ScrollView, Prefab, instantiate, Toggle } from 'cc';
const { ccclass, property } = _decorator;

import { LeaderboardSystem, LeaderboardEntry, RankType } from '../core/LeaderboardSystem';
import { UIManager } from './UIManager';

/**
 * LeaderboardUI - 排行榜界面
 * 支持战力榜/胜场榜切换
 */
@ccclass('LeaderboardUI')
export class LeaderboardUI extends Component {
    @property(ScrollView)
    rankScroll: ScrollView | null = null;

    @property(Prefab)
    rankItemPrefab: Prefab | null = null;

    @property(Button)
    powerTabBtn: Button | null = null;

    @property(Button)
    winsTabBtn: Button | null = null;

    @property(Button)
    backBtn: Button | null = null;

    @property(Label)
    titleLabel: Label | null = null;

    private currentType: RankType = 'power';

    onLoad() {
        this.powerTabBtn?.node.on('click', () => this.switchTab('power'), this);
        this.winsTabBtn?.node.on('click', () => this.switchTab('wins'), this);
        this.backBtn?.node.on('click', this.onBack, this);
    }

    public async onShow(): Promise<void> {
        await this.loadLeaderboard();
    }

    private switchTab(type: RankType): void {
        this.currentType = type;
        this.loadLeaderboard();
    }

    private async loadLeaderboard(): Promise<void> {
        if (this.titleLabel) {
            this.titleLabel.string = this.currentType === 'power' ? '战力排行榜' : '胜场排行榜';
        }

        const system = LeaderboardSystem.getInstance();
        const entries = await system.getLeaderboard(this.currentType, 50);

        const content = this.rankScroll?.content;
        if (!content) return;
        content.removeAllChildren();

        for (const entry of entries) {
            this.createRankItem(content, entry);
        }
    }

    private createRankItem(parent: Node, entry: LeaderboardEntry): void {
        let itemNode: Node;
        if (this.rankItemPrefab) {
            itemNode = instantiate(this.rankItemPrefab);
        } else {
            itemNode = new Node('RankItem');
        }
        parent.addChild(itemNode);

        const label = itemNode.getComponentInChildren(Label);
        if (label) {
            const rankIcon = entry.rank <= 3 ? ['🥇', '🥈', '🥉'][entry.rank - 1] : `${entry.rank}.`;
            const scoreText = this.currentType === 'power'
                ? `战力:${entry.score}`
                : `${entry.score}胜`;
            label.string = `${rankIcon} ${entry.nickname} Lv.${entry.level} | ${scoreText}`;
        }
    }

    private onBack(): void {
        UIManager.instance?.closePanel();
    }
}
