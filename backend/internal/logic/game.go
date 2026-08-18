package logic

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"automatictools/backend/internal/store"

	"gorm.io/gorm"
)

// ===== DTO Types =====

type GamePlayerDTO struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"userId"`
	Nickname    string `json:"nickname"`
	Level       int    `json:"level"`
	Exp         int64  `json:"exp"`
	Gold        int64  `json:"gold"`
	CombatPower int64  `json:"combatPower"`
	Wins        int    `json:"wins"`
	Losses      int    `json:"losses"`
}

type EquipmentDTO struct {
	ID         int64    `json:"id"`
	TemplateID int64    `json:"templateId"`
	Name       string   `json:"name"`
	Slot       string   `json:"slot"`
	Rarity     int      `json:"rarity"`
	Atk        int      `json:"atk"`
	Def        int      `json:"def"`
	Hp         int      `json:"hp"`
	Spd        int      `json:"spd"`
	BonusAttrs []BonusAttr `json:"bonusAttrs"`
	Equipped   bool     `json:"equipped"`
	IconPath   string   `json:"iconPath"`
}

type BonusAttr struct {
	Attr  string `json:"attr"`
	Value int    `json:"value"`
}

type BoxResultDTO struct {
	Equipment EquipmentDTO `json:"equipment"`
	IsPity    bool         `json:"isPity"`
}

type BattleResultDTO struct {
	WinnerID   int64      `json:"winnerId"`
	BattleLog  []BattleRound `json:"battleLog"`
	RewardGold int64      `json:"rewardGold"`
	RewardExp  int64      `json:"rewardExp"`
	IsWin      bool       `json:"isWin"`
}

type BattleRound struct {
	Round       int   `json:"round"`
	AttackerDmg int   `json:"attackerDmg"`
	DefenderDmg int   `json:"defenderDmg"`
	AttackerHp  int   `json:"attackerHp"`
	DefenderHp  int   `json:"defenderHp"`
}

type OpponentDTO struct {
	PlayerID    int64  `json:"playerId"`
	Nickname    string `json:"nickname"`
	Level       int    `json:"level"`
	CombatPower int64  `json:"combatPower"`
	Wins        int    `json:"wins"`
}

type LeaderboardEntryDTO struct {
	Rank        int    `json:"rank"`
	PlayerID    int64  `json:"playerId"`
	Nickname    string `json:"nickname"`
	Level       int    `json:"level"`
	Score       int64  `json:"score"`
	CombatPower int64  `json:"combatPower,omitempty"`
	Wins        int    `json:"wins,omitempty"`
}

// ===== Constants =====

const (
	maxDailyChallenges = 10
	pityThreshold      = 50
)

// rarity probabilities (cumulative)
var rarityWeights = []struct {
	Rarity int
	Weight int
}{
	{1, 50}, // common white
	{2, 25}, // uncommon green
	{3, 15}, // rare blue
	{4, 8},  // epic purple
	{5, 2},  // legendary orange
}

// ===== Player Management =====

func (a *Service) GetOrCreateGamePlayer(ctx context.Context, userID int64) (GamePlayerDTO, error) {
	var player store.GamePlayer
	err := a.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Take(&player).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := unixNow()
		player = store.GamePlayer{
			UserID:    userID,
			Nickname:  "冒险者",
			Level:     1,
			Exp:       0,
			Gold:      100, // initial gold
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := a.db.WithContext(ctx).Create(&player).Error; err != nil {
			return GamePlayerDTO{}, err
		}
	} else if err != nil {
		return GamePlayerDTO{}, err
	}

	return gamePlayerDTO(player), nil
}

func gamePlayerDTO(p store.GamePlayer) GamePlayerDTO {
	return GamePlayerDTO{
		ID:          p.ID,
		UserID:      p.UserID,
		Nickname:    p.Nickname,
		Level:       p.Level,
		Exp:         p.Exp,
		Gold:        p.Gold,
		CombatPower: p.CombatPower,
		Wins:        p.Wins,
		Losses:      p.Losses,
	}
}

// ===== Box Opening =====

func (a *Service) OpenBox(ctx context.Context, userID int64) (BoxResultDTO, error) {
	var player store.GamePlayer
	err := a.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Take(&player).Error
	if err != nil {
		return BoxResultDTO{}, unauthorized("请先创建角色。")
	}

	// Determine rarity
	rarity := a.rollRarity(player.BoxPityCounter)
	isPity := false
	if player.BoxPityCounter+1 >= pityThreshold && rarity < 4 {
		rarity = 4
		isPity = true
	}

	// Pick a random template of this rarity
	var template store.GameEquipmentTemplate
	err = a.db.WithContext(ctx).
		Where("rarity = ?", rarity).
		Order("RANDOM()").
		Limit(1).
		Take(&template).Error
	if err != nil {
		return BoxResultDTO{}, errors.New("没有可用的装备模板")
	}

	// Generate equipment instance with random bonus
	equipment := a.generateEquipment(player.ID, &template)

	// Save in transaction
	var resultEquip store.GamePlayerEquipment
	err = a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&equipment).Error; err != nil {
			return err
		}
		resultEquip = equipment

		// Update pity counter
		newPity := 0
		if rarity < 4 {
			newPity = player.BoxPityCounter + 1
		}
		if err := tx.Model(&store.GamePlayer{}).
			Where("id = ?", player.ID).
			Updates(map[string]any{
				"box_pity_counter": newPity,
				"updated_at":       unixNow(),
			}).Error; err != nil {
			return err
		}

		// Record box opening
		record := store.GameBoxRecord{
			PlayerID:          player.ID,
			BoxType:           "normal",
			ResultEquipmentID: &equipment.ID,
			IsPity:            isPity,
			CreatedAt:         unixNow(),
		}
		return tx.Create(&record).Error
	})
	if err != nil {
		return BoxResultDTO{}, err
	}

	return BoxResultDTO{
		Equipment: equipmentDTOFromInstance(&resultEquip, &template),
		IsPity:    isPity,
	}, nil
}

func (a *Service) rollRarity(pityCounter int) int {
	total := 0
	for _, w := range rarityWeights {
		total += w.Weight
	}
	roll := rand.Intn(total)
	cumulative := 0
	for _, w := range rarityWeights {
		cumulative += w.Weight
		if roll < cumulative {
			return w.Rarity
		}
	}
	return 1
}

func (a *Service) generateEquipment(playerID int64, template *store.GameEquipmentTemplate) store.GamePlayerEquipment {
	// Random variance ±20% on base stats
	variance := func(base int) int {
		if base == 0 {
			return 0
		}
		delta := float64(base) * 0.2
		return base + int(rand.Float64()*2*delta-delta)
	}

	// Generate 0-3 bonus attributes based on rarity
	bonusCount := rand.Intn(template.Rarity) // 0 to rarity-1
	if bonusCount > 3 {
		bonusCount = 3
	}
	bonusAttrs := make([]BonusAttr, 0, bonusCount)
	attrPool := []string{"atk", "def", "hp", "spd", "crit", "dodge"}
	for i := 0; i < bonusCount; i++ {
		attr := attrPool[rand.Intn(len(attrPool))]
		value := rand.Intn(5*template.Rarity) + 1
		bonusAttrs = append(bonusAttrs, BonusAttr{Attr: attr, Value: value})
	}
	bonusJSON, _ := json.Marshal(bonusAttrs)

	return store.GamePlayerEquipment{
		PlayerID:   playerID,
		TemplateID: template.ID,
		Rarity:     template.Rarity,
		Atk:        variance(template.BaseAtk),
		Def:        variance(template.BaseDef),
		Hp:         variance(template.BaseHp),
		Spd:        variance(template.BaseSpd),
		BonusAttrs: string(bonusJSON),
		Equipped:   false,
		ObtainedAt: unixNow(),
	}
}

// ===== Equipment Management =====

func (a *Service) ListPlayerEquipments(ctx context.Context, userID int64) ([]EquipmentDTO, error) {
	var player store.GamePlayer
	if err := a.db.WithContext(ctx).Where("user_id = ?", userID).Take(&player).Error; err != nil {
		return nil, unauthorized("请先创建角色。")
	}

	var equipments []store.GamePlayerEquipment
	if err := a.db.WithContext(ctx).
		Where("player_id = ?", player.ID).
		Order("rarity DESC, obtained_at DESC").
		Find(&equipments).Error; err != nil {
		return nil, err
	}

	// Fetch templates for names/icons
	templateIDs := make([]int64, 0, len(equipments))
	for _, e := range equipments {
		templateIDs = append(templateIDs, e.TemplateID)
	}
	var templates []store.GameEquipmentTemplate
	if len(templateIDs) > 0 {
		a.db.WithContext(ctx).Where("id IN ?", templateIDs).Find(&templates)
	}
	templateMap := make(map[int64]*store.GameEquipmentTemplate, len(templates))
	for i := range templates {
		templateMap[templates[i].ID] = &templates[i]
	}

	result := make([]EquipmentDTO, 0, len(equipments))
	for i := range equipments {
		tmpl := templateMap[equipments[i].TemplateID]
		result = append(result, equipmentDTOFromInstance(&equipments[i], tmpl))
	}
	return result, nil
}

type EquipRequest struct {
	EquipmentID int64 `json:"equipmentId"`
}

func (a *Service) EquipItem(ctx context.Context, userID int64, req EquipRequest) error {
	var player store.GamePlayer
	if err := a.db.WithContext(ctx).Where("user_id = ?", userID).Take(&player).Error; err != nil {
		return unauthorized("请先创建角色。")
	}

	var equipment store.GamePlayerEquipment
	if err := a.db.WithContext(ctx).
		Where("id = ? AND player_id = ?", req.EquipmentID, player.ID).
		Take(&equipment).Error; err != nil {
		return badRequest("装备不存在。")
	}

	// Get template to know slot
	var template store.GameEquipmentTemplate
	if err := a.db.WithContext(ctx).Where("id = ?", equipment.TemplateID).Take(&template).Error; err != nil {
		return err
	}

	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unequip current item in same slot
		if err := tx.Model(&store.GamePlayerEquipment{}).
			Where("player_id = ? AND equipped = ? AND template_id IN (SELECT id FROM game_equipment_templates WHERE slot = ?)",
				player.ID, true, template.Slot).
			Update("equipped", false).Error; err != nil {
			return err
		}
		// Equip new item
		if err := tx.Model(&store.GamePlayerEquipment{}).
			Where("id = ?", equipment.ID).
			Update("equipped", true).Error; err != nil {
			return err
		}
		// Recalculate combat power
		return a.recalcCombatPower(ctx, tx, player.ID)
	})
}

func (a *Service) recalcCombatPower(ctx context.Context, tx *gorm.DB, playerID int64) error {
	var equipments []store.GamePlayerEquipment
	if err := tx.Where("player_id = ? AND equipped = ?", playerID, true).Find(&equipments).Error; err != nil {
		return err
	}

	var player store.GamePlayer
	if err := tx.Where("id = ?", playerID).Take(&player).Error; err != nil {
		return err
	}

	totalAtk, totalDef, totalHp, totalSpd := 0, 0, 0, 0
	for _, e := range equipments {
		totalAtk += e.Atk
		totalDef += e.Def
		totalHp += e.Hp
		totalSpd += e.Spd
	}

	// Base stats from level
	baseAtk := 10 + player.Level*2
	baseDef := 8 + player.Level*2
	baseHp := 100 + player.Level*15
	baseSpd := 10 + player.Level

	finalAtk := baseAtk + totalAtk
	finalDef := baseDef + totalDef
	finalHp := baseHp + totalHp
	finalSpd := baseSpd + totalSpd

	// Combat power formula
	power := int64(float64(finalAtk)*1.5 + float64(finalDef)*1.2 + float64(finalHp)*0.1 + float64(finalSpd)*1.0)

	return tx.Model(&store.GamePlayer{}).
		Where("id = ?", playerID).
		Updates(map[string]any{
			"combat_power": power,
			"updated_at":   unixNow(),
		}).Error
}

// ===== Battle System =====

func (a *Service) GetOpponents(ctx context.Context, userID int64) ([]OpponentDTO, error) {
	var player store.GamePlayer
	if err := a.db.WithContext(ctx).Where("user_id = ?", userID).Take(&player).Error; err != nil {
		return nil, unauthorized("请先创建角色。")
	}

	// Reset daily challenge count if new day
	today := time.Now().Format("2006-01-02")
	if player.DailyChallengeDate != today {
		a.db.WithContext(ctx).Model(&store.GamePlayer{}).
			Where("id = ?", player.ID).
			Updates(map[string]any{
				"daily_challenge_count": 0,
				"daily_challenge_date":  today,
			})
		player.DailyChallengeCount = 0
	}

	// Find opponents within ±10% combat power range
	powerRange := float64(player.CombatPower) * 0.1
	minPower := int64(float64(player.CombatPower) - powerRange)
	maxPower := int64(float64(player.CombatPower) + powerRange)
	if minPower < 0 {
		minPower = 0
	}

	var opponents []store.GamePlayer
	err := a.db.WithContext(ctx).
		Where("id != ? AND combat_power BETWEEN ? AND ?", player.ID, minPower, maxPower).
		Order("RANDOM()").
		Limit(5).
		Find(&opponents).Error
	if err != nil {
		return nil, err
	}

	result := make([]OpponentDTO, 0, len(opponents))
	for _, opp := range opponents {
		result = append(result, OpponentDTO{
			PlayerID:    opp.ID,
			Nickname:    opp.Nickname,
			Level:       opp.Level,
			CombatPower: opp.CombatPower,
			Wins:        opp.Wins,
		})
	}
	return result, nil
}

type ChallengeRequest struct {
	DefenderID int64 `json:"defenderId"`
}

func (a *Service) Challenge(ctx context.Context, userID int64, req ChallengeRequest) (BattleResultDTO, error) {
	var attacker store.GamePlayer
	if err := a.db.WithContext(ctx).Where("user_id = ?", userID).Take(&attacker).Error; err != nil {
		return BattleResultDTO{}, unauthorized("请先创建角色。")
	}

	// Check daily limit
	today := time.Now().Format("2006-01-02")
	if attacker.DailyChallengeDate != today {
		attacker.DailyChallengeCount = 0
	}
	if attacker.DailyChallengeCount >= maxDailyChallenges {
		return BattleResultDTO{}, badRequest("今日挑战次数已用完。")
	}

	var defender store.GamePlayer
	if err := a.db.WithContext(ctx).Where("id = ?", req.DefenderID).Take(&defender).Error; err != nil {
		return BattleResultDTO{}, badRequest("对手不存在。")
	}

	// Simulate battle
	battleLog, winnerID := a.simulateBattle(&attacker, &defender)
	isWin := winnerID == attacker.ID

	// Calculate rewards
	var rewardGold, rewardExp int64
	if isWin {
		rewardGold = 50 + int64(rand.Intn(50)) + int64(defender.Level)*5
		rewardExp = 30 + int64(rand.Intn(20)) + int64(defender.Level)*3
	} else {
		rewardGold = 10
		rewardExp = 10
	}

	// Save battle result
	battle := store.GameBattle{
		AttackerID: attacker.ID,
		DefenderID: defender.ID,
		WinnerID:   winnerID,
		BattleLog:  mustJSON(battleLog),
		RewardGold: rewardGold,
		RewardExp:  rewardExp,
		CreatedAt:  unixNow(),
	}

	err := a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&battle).Error; err != nil {
			return err
		}

		// Update attacker stats
		attackerUpdates := map[string]any{
			"gold":                  attacker.Gold + rewardGold,
			"exp":                   attacker.Exp + rewardExp,
			"daily_challenge_count": attacker.DailyChallengeCount + 1,
			"daily_challenge_date":  today,
			"updated_at":            unixNow(),
		}
		if isWin {
			attackerUpdates["wins"] = attacker.Wins + 1
		} else {
			attackerUpdates["losses"] = attacker.Losses + 1
		}
		if err := tx.Model(&store.GamePlayer{}).Where("id = ?", attacker.ID).Updates(attackerUpdates).Error; err != nil {
			return err
		}

		// Update defender stats
		defenderUpdates := map[string]any{"updated_at": unixNow()}
		if isWin {
			defenderUpdates["losses"] = defender.Losses + 1
		} else {
			defenderUpdates["wins"] = defender.Wins + 1
		}
		return tx.Model(&store.GamePlayer{}).Where("id = ?", defender.ID).Updates(defenderUpdates).Error
	})
	if err != nil {
		return BattleResultDTO{}, err
	}

	return BattleResultDTO{
		WinnerID:   winnerID,
		BattleLog:  battleLog,
		RewardGold: rewardGold,
		RewardExp:  rewardExp,
		IsWin:      isWin,
	}, nil
}

func (a *Service) simulateBattle(attacker, defender *store.GamePlayer) ([]BattleRound, int64) {
	// Get effective stats
	atkStats := a.getEffectiveStats(attacker)
	defStats := a.getEffectiveStats(defender)

	attackerHp := atkStats.Hp
	defenderHp := defStats.Hp
	rounds := make([]BattleRound, 0, 20)

	for round := 1; round <= 20; round++ {
		// Attacker attacks
		atkDmg := calcDamage(atkStats.Atk, defStats.Def)
		defenderHp -= atkDmg

		// Defender attacks
		defDmg := calcDamage(defStats.Atk, atkStats.Def)
		attackerHp -= defDmg

		rounds = append(rounds, BattleRound{
			Round:       round,
			AttackerDmg: atkDmg,
			DefenderDmg: defDmg,
			AttackerHp:  max(0, attackerHp),
			DefenderHp:  max(0, defenderHp),
		})

		if defenderHp <= 0 {
			return rounds, attacker.ID
		}
		if attackerHp <= 0 {
			return rounds, defender.ID
		}
	}

	// Timeout: higher remaining HP wins
	if attackerHp >= defenderHp {
		return rounds, attacker.ID
	}
	return rounds, defender.ID
}

type effectiveStats struct {
	Atk, Def, Hp, Spd int
}

func (a *Service) getEffectiveStats(player *store.GamePlayer) effectiveStats {
	var equipments []store.GamePlayerEquipment
	a.db.Where("player_id = ? AND equipped = ?", player.ID, true).Find(&equipments)

	totalAtk, totalDef, totalHp, totalSpd := 0, 0, 0, 0
	for _, e := range equipments {
		totalAtk += e.Atk
		totalDef += e.Def
		totalHp += e.Hp
		totalSpd += e.Spd
	}

	return effectiveStats{
		Atk: 10 + player.Level*2 + totalAtk,
		Def: 8 + player.Level*2 + totalDef,
		Hp:  100 + player.Level*15 + totalHp,
		Spd: 10 + player.Level + totalSpd,
	}
}

func calcDamage(atk, def int) int {
	base := atk - def/2
	if base < 1 {
		base = 1
	}
	// ±15% random variance
	variance := float64(base) * 0.15
	dmg := float64(base) + (rand.Float64()*2-1)*variance
	if dmg < 1 {
		dmg = 1
	}
	return int(dmg)
}

// ===== Leaderboard =====

func (a *Service) GetLeaderboard(ctx context.Context, rankType string, limit int) ([]LeaderboardEntryDTO, error) {
	if rankType != "power" && rankType != "wins" {
		rankType = "power"
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var players []store.GamePlayer
	orderClause := "combat_power DESC"
	if rankType == "wins" {
		orderClause = "wins DESC"
	}

	if err := a.db.WithContext(ctx).
		Order(orderClause).
		Limit(limit).
		Find(&players).Error; err != nil {
		return nil, err
	}

	result := make([]LeaderboardEntryDTO, 0, len(players))
	for i, p := range players {
		entry := LeaderboardEntryDTO{
			Rank:        i + 1,
			PlayerID:    p.ID,
			Nickname:    p.Nickname,
			Level:       p.Level,
			CombatPower: p.CombatPower,
			Wins:        p.Wins,
		}
		if rankType == "power" {
			entry.Score = p.CombatPower
		} else {
			entry.Score = int64(p.Wins)
		}
		result = append(result, entry)
	}
	return result, nil
}

// ===== Helpers =====

func equipmentDTOFromInstance(e *store.GamePlayerEquipment, tmpl *store.GameEquipmentTemplate) EquipmentDTO {
	dto := EquipmentDTO{
		ID:         e.ID,
		TemplateID: e.TemplateID,
		Rarity:     e.Rarity,
		Atk:        e.Atk,
		Def:        e.Def,
		Hp:         e.Hp,
		Spd:        e.Spd,
		Equipped:   e.Equipped,
	}
	if tmpl != nil {
		dto.Name = tmpl.Name
		dto.Slot = tmpl.Slot
		dto.IconPath = tmpl.IconPath
	}
	var bonusAttrs []BonusAttr
	if err := json.Unmarshal([]byte(e.BonusAttrs), &bonusAttrs); err == nil {
		dto.BonusAttrs = bonusAttrs
	} else {
		dto.BonusAttrs = []BonusAttr{}
	}
	return dto
}

func mustJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
