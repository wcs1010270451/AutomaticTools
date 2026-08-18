import { _decorator, Component, Node, instantiate, Prefab, resources, tween, Vec3, UIOpacity } from 'cc';
const { ccclass, property } = _decorator;

/**
 * UIManager - UI 管理器
 * 代码驱动 UI 面板的加载、显示、隐藏、切换
 * 采用栈式管理，支持多层 UI 叠加
 */

export enum UI_PANEL {
    MAIN = 'MainUI',
    BOX_OPEN = 'BoxOpenUI',
    EQUIPMENT = 'EquipmentUI',
    BATTLE = 'BattleUI',
    LEADERBOARD = 'LeaderboardUI',
    LOGIN = 'LoginUI',
}

@ccclass('UIManager')
export class UIManager extends Component {
    private static _instance: UIManager;

    public static get instance(): UIManager {
        return UIManager._instance;
    }

    /** UI 面板预制体路径前缀 */
    private prefabPath: string = 'prefabs/ui/';

    /** 已加载的面板缓存 */
    private panelCache: Map<string, Node> = new Map();

    /** 当前显示的面板栈 */
    private panelStack: string[] = [];

    /** UI 根节点 (Canvas) */
    @property(Node)
    uiRoot: Node | null = null;

    onLoad() {
        UIManager._instance = this;
    }

    /**
     * 打开一个 UI 面板
     * @param panelName 面板名称
     * @param params 传递给面板的参数
     */
    public async openPanel(panelName: UI_PANEL, params?: any): Promise<void> {
        // 如果已在栈顶，不重复打开
        if (this.panelStack[this.panelStack.length - 1] === panelName) {
            return;
        }

        let panelNode = this.panelCache.get(panelName);

        if (!panelNode) {
            // 从资源加载预制体
            panelNode = await this.loadPanel(panelName);
            if (!panelNode) return;
        }

        // 隐藏当前栈顶面板
        this.hideTopPanel();

        // 显示新面板
        panelNode.active = true;
        this.panelStack.push(panelName);

        // 入场动画
        this.playOpenAnimation(panelNode);

        // 调用面板的 onShow 方法
        const components = panelNode.components;
        for (const comp of components) {
            if ('onShow' in comp && typeof (comp as any).onShow === 'function') {
                (comp as any).onShow(params);
            }
        }
    }

    /**
     * 关闭当前面板，返回上一级
     */
    public closePanel(): void {
        if (this.panelStack.length === 0) return;

        const topPanel = this.panelStack.pop()!;
        const panelNode = this.panelCache.get(topPanel);

        if (panelNode) {
            // 调用面板的 onHide 方法
            const components = panelNode.components;
            for (const comp of components) {
                if ('onHide' in comp && typeof (comp as any).onHide === 'function') {
                    (comp as any).onHide();
                }
            }

            // 出场动画后隐藏
            this.playCloseAnimation(panelNode, () => {
                panelNode.active = false;
            });
        }

        // 显示下一个面板
        this.showTopPanel();
    }

    /**
     * 关闭所有面板并打开指定面板
     */
    public switchPanel(panelName: UI_PANEL, params?: any): void {
        // 隐藏所有面板
        for (const name of this.panelStack) {
            const node = this.panelCache.get(name);
            if (node) node.active = false;
        }
        this.panelStack = [];

        // 打开目标面板
        this.openPanel(panelName, params);
    }

    /**
     * 获取当前栈顶面板名称
     */
    public getCurrentPanel(): string | null {
        return this.panelStack.length > 0 ? this.panelStack[this.panelStack.length - 1] : null;
    }

    private async loadPanel(panelName: UI_PANEL): Promise<Node | null> {
        return new Promise((resolve) => {
            const path = `${this.prefabPath}${panelName}`;
            resources.load(path, Prefab, (err, prefab) => {
                if (err) {
                    console.error(`加载 UI 预制体失败: ${path}`, err);
                    resolve(null);
                    return;
                }
                const node = instantiate(prefab);
                if (this.uiRoot) {
                    this.uiRoot.addChild(node);
                }
                this.panelCache.set(panelName, node);
                resolve(node);
            });
        });
    }

    private hideTopPanel(): void {
        if (this.panelStack.length > 0) {
            const topName = this.panelStack[this.panelStack.length - 1];
            const topNode = this.panelCache.get(topName);
            if (topNode) topNode.active = false;
        }
    }

    private showTopPanel(): void {
        if (this.panelStack.length > 0) {
            const topName = this.panelStack[this.panelStack.length - 1];
            const topNode = this.panelCache.get(topName);
            if (topNode) topNode.active = true;
        }
    }

    private playOpenAnimation(node: Node): void {
        node.setScale(new Vec3(0.8, 0.8, 1));
        let opacity = node.getComponent(UIOpacity);
        if (!opacity) {
            opacity = node.addComponent(UIOpacity);
        }
        opacity.opacity = 0;

        tween(node)
            .to(0.2, { scale: new Vec3(1, 1, 1) }, { easing: 'backOut' })
            .start();
        tween(opacity)
            .to(0.15, { opacity: 255 })
            .start();
    }

    private playCloseAnimation(node: Node, onComplete: () => void): void {
        tween(node)
            .to(0.15, { scale: new Vec3(0.9, 0.9, 1) })
            .call(onComplete)
            .start();
    }
}
