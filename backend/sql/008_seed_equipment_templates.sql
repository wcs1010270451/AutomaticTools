-- 装备模板种子数据
-- 用于初始化 game_equipment_templates 表
-- 5个槽位 x 5个稀有度 x 2-3种 = 约60件装备

BEGIN;

-- ===== 武器 (weapon) =====
-- 白色 (rarity=1)
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('木剑', 'weapon', 1, 8, 0, 0, 2, 'textures/equip/weapon_wood_sword', '新手冒险者的第一把武器', 1723420800),
('石斧', 'weapon', 1, 10, 0, 0, 1, 'textures/equip/weapon_stone_axe', '笨重但还算结实', 1723420800);

-- 绿色 (rarity=2)
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('铁剑', 'weapon', 2, 18, 0, 0, 3, 'textures/equip/weapon_iron_sword', '标准骑士佩剑', 1723420800),
('猎弓', 'weapon', 2, 15, 0, 0, 6, 'textures/equip/weapon_hunt_bow', '轻便的短弓', 1723420800);

-- 蓝色 (rarity=3)
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('精钢长剑', 'weapon', 3, 32, 0, 0, 5, 'textures/equip/weapon_steel_sword', '由精钢锻造的利刃', 1723420800),
('魔法杖', 'weapon', 3, 28, 0, 10, 4, 'textures/equip/weapon_magic_staff', '蕴含魔力的法杖', 1723420800);

-- 紫色 (rarity=4)
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('烈焰之刃', 'weapon', 4, 55, 0, 0, 8, 'textures/equip/weapon_flame_blade', '附着永恒烈焰的魔剑', 1723420800),
('暗影匕首', 'weapon', 4, 48, 0, 0, 14, 'textures/equip/weapon_shadow_dagger', '来自深渊的致命短刃', 1723420800);

-- 橙色 (rarity=5)
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('圣光裁决者', 'weapon', 5, 90, 5, 20, 12, 'textures/equip/weapon_holy_judge', '传说中制裁邪恶的神器', 1723420800),
('龙牙巨剑', 'weapon', 5, 100, 0, 0, 8, 'textures/equip/weapon_dragon_fang', '由远古巨龙之牙铸成', 1723420800);

-- ===== 头盔 (helmet) =====
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('布帽', 'helmet', 1, 0, 3, 10, 0, 'textures/equip/helmet_cloth', '勉强能遮阳', 1723420800),
('皮盔', 'helmet', 1, 0, 5, 15, 0, 'textures/equip/helmet_leather', '简陋的皮革头盔', 1723420800),
('铁盔', 'helmet', 2, 0, 10, 25, 0, 'textures/equip/helmet_iron', '标准士兵头盔', 1723420800),
('角盔', 'helmet', 2, 2, 12, 20, 0, 'textures/equip/helmet_horn', '带有犄角的战盔', 1723420800),
('骑士面甲', 'helmet', 3, 0, 22, 40, 0, 'textures/equip/helmet_knight', '精工打造的全覆面甲', 1723420800),
('智慧之冠', 'helmet', 3, 5, 15, 35, 3, 'textures/equip/helmet_wisdom', '增加思维敏捷的头冠', 1723420800),
('龙角战盔', 'helmet', 4, 5, 38, 60, 3, 'textures/equip/helmet_dragon_horn', '以龙角装饰的传奇战盔', 1723420800),
('暗影兜帽', 'helmet', 4, 8, 30, 45, 8, 'textures/equip/helmet_shadow_hood', '隐匿于暗影中的兜帽', 1723420800),
('王者之冠', 'helmet', 5, 10, 55, 90, 5, 'textures/equip/helmet_king_crown', '古代帝王的神圣王冠', 1723420800);

-- ===== 铠甲 (armor) =====
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('布衣', 'armor', 1, 0, 5, 20, 0, 'textures/equip/armor_cloth', '普通旅行者的衣物', 1723420800),
('皮甲', 'armor', 1, 0, 8, 30, 0, 'textures/equip/armor_leather', '基础皮革护甲', 1723420800),
('锁子甲', 'armor', 2, 0, 16, 50, -1, 'textures/equip/armor_chain', '由铁环编织的铠甲', 1723420800),
('轻皮甲', 'armor', 2, 0, 12, 40, 3, 'textures/equip/armor_light_leather', '轻便灵活的皮甲', 1723420800),
('板甲', 'armor', 3, 0, 30, 80, -2, 'textures/equip/armor_plate', '厚重但防御出色的全身甲', 1723420800),
('游侠之衣', 'armor', 3, 3, 22, 55, 5, 'textures/equip/armor_ranger', '森林游侠的轻巧装束', 1723420800),
('烈焰战甲', 'armor', 4, 5, 50, 100, 0, 'textures/equip/armor_flame', '被火焰附魔的重甲', 1723420800),
('暗影轻甲', 'armor', 4, 8, 38, 70, 8, 'textures/equip/armor_shadow', '暗夜刺客的贴身轻甲', 1723420800),
('神圣铠甲', 'armor', 5, 10, 75, 150, 3, 'textures/equip/armor_holy', '受到神明庇佑的不朽铠甲', 1723420800);

-- ===== 盾牌 (shield) =====
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('木盾', 'shield', 1, 0, 6, 15, 0, 'textures/equip/shield_wood', '简陋的木制圆盾', 1723420800),
('铁盾', 'shield', 2, 0, 14, 30, 0, 'textures/equip/shield_iron', '标准步兵盾牌', 1723420800),
('塔盾', 'shield', 3, 0, 28, 55, -2, 'textures/equip/shield_tower', '能遮蔽全身的大型塔盾', 1723420800),
('龙鳞盾', 'shield', 4, 0, 45, 80, 0, 'textures/equip/shield_dragon_scale', '以龙鳞打造的坚固盾牌', 1723420800),
('圣光壁垒', 'shield', 5, 0, 70, 120, 2, 'textures/equip/shield_holy_barrier', '凝聚圣光之力的终极防御', 1723420800);

-- ===== 饰品 (accessory) =====
INSERT INTO game_equipment_templates (name, slot, rarity, base_atk, base_def, base_hp, base_spd, icon_path, description, created_at)
VALUES
('铜戒指', 'accessory', 1, 2, 2, 5, 1, 'textures/equip/acc_copper_ring', '普通的铜质戒指', 1723420800),
('翡翠坠', 'accessory', 2, 4, 4, 15, 2, 'textures/equip/acc_jade_pendant', '温润的翡翠吊坠', 1723420800),
('红宝石戒', 'accessory', 3, 10, 5, 20, 4, 'textures/equip/acc_ruby_ring', '镶嵌红宝石的魔法戒指', 1723420800),
('龙血项链', 'accessory', 4, 15, 12, 40, 6, 'textures/equip/acc_dragon_blood_necklace', '蕴含龙血之力的项链', 1723420800),
('永恒之心', 'accessory', 5, 25, 20, 60, 10, 'textures/equip/acc_eternal_heart', '传说中永不熄灭的生命宝石', 1723420800);

COMMIT;
