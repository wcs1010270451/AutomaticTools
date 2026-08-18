import { _decorator, Component, Node, resources, JsonAsset } from 'cc';
const { ccclass, property } = _decorator;

import { ApiClient } from './network/ApiClient';
import { AuthService } from './network/AuthService';
import { PlayerData } from './core/PlayerData';
import { UIManager, UI_PANEL } from './ui/UIManager';

/**
 * GameInit - 游戏初始化入口
 * 挂载在场景根节点，负责整个游戏的启动流程
 */
@ccclass('GameInit')
export class GameInit extends Component {
    @property(Node)
    uiRoot: Node | null = null;

    async onLoad() {
        console.log('[GameInit] 游戏启动...');

        // 1. 加载配置
        const config = await this.loadConfig();
        const apiBaseUrl = config?.game?.apiBaseUrl || 'http://localhost:8080';

        // 2. 初始化 API 客户端
        ApiClient.getInstance().init(apiBaseUrl);

        // 3. 尝试恢复登录状态
        const auth = AuthService.getInstance();
        const hasSession = auth.restoreSession();

        if (hasSession) {
            // 已登录，直接初始化玩家数据
            await this.initGame();
        } else {
            // 未登录，显示登录界面或自动登录
            await this.handleLogin();
        }
    }

    private async loadConfig(): Promise<any> {
        return new Promise((resolve) => {
            resources.load('configs/game_config', JsonAsset, (err, jsonAsset) => {
                if (err) {
                    console.warn('[GameInit] 配置文件加载失败，使用默认配置');
                    resolve(null);
                    return;
                }
                resolve(jsonAsset.json);
            });
        });
    }

    private async handleLogin(): Promise<void> {
        // @ts-ignore - 检测是否在微信小游戏环境
        if (typeof wx !== 'undefined') {
            // 微信环境自动登录
            const result = await AuthService.getInstance().loginWithWechat();
            if (result) {
                AuthService.getInstance().saveToken(result.token);
                await this.initGame();
                return;
            }
        }

        // H5 环境：简化处理，先用游客模式
        // TODO: 后续接入邮箱/手机号登录界面
        console.log('[GameInit] H5 环境，等待登录...');
        UIManager.instance?.openPanel(UI_PANEL.LOGIN);
    }

    private async initGame(): Promise<void> {
        // 初始化玩家数据
        const player = await PlayerData.getInstance().init();
        if (!player) {
            console.error('[GameInit] 玩家数据初始化失败');
            return;
        }

        console.log(`[GameInit] 欢迎回来, ${player.nickname}!`);

        // 打开主界面
        UIManager.instance?.switchPanel(UI_PANEL.MAIN);
    }
}
