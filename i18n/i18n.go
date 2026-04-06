package i18n

import "fmt"

// Lang represents a supported language.
type Lang int

const (
	EN Lang = iota
	ZH
	LangCount // sentinel: total number of languages
)

var current Lang = EN

// SetLang sets the current language.
func SetLang(l Lang) { current = l }

// GetLang returns the current language.
func GetLang() Lang { return current }

// LangName returns the display name for a language (always in its native form).
func LangName(l Lang) string {
	switch l {
	case ZH:
		return "简体中文"
	default:
		return "English"
	}
}

// T translates a string key to the current language.
// For English, the key itself IS the English text (returned as-is).
// For other languages, looks up the translation map; falls back to key if not found.
func T(key string) string {
	if current == EN {
		return key
	}
	if v, ok := zhCN[key]; ok {
		return v
	}
	return key
}

// Tf translates a format string then applies fmt.Sprintf with the given arguments.
func Tf(format string, args ...interface{}) string {
	return fmt.Sprintf(T(format), args...)
}

// HelpText returns the full help screen text for the current language.
func HelpText() string {
	if current == ZH {
		return zhHelpText
	}
	return enHelpText
}

const enHelpText = `
CIV-TUI HELP
============

MOVEMENT:
  Arrow keys / hjkl - Move cursor / selected unit
  Green tiles show reachable positions

UNIT ACTIONS:
  F - Found City (Settler only)
  W - Wait/Skip unit turn
  N - Select next unit with moves
  R - Ranged attack mode (Archers)
  I - Build improvement (Worker)

CITY ACTIONS:
  B - Open build menu (when on city)
  Enter (on own city, no unit) - City details

INSPECT:
  V - View tile details

RESEARCH:
  T - Open tech research menu

DIPLOMACY:
  D - Open diplomacy menu

GAME:
  S - Save game
  Enter - End turn

DISPLAY:
  Blue units/cities = yours
  Red units/cities = enemy
  Green highlight = reachable tiles
  Red highlight = ranged attack range
  Dimmed = explored but not visible
  Dark tiles = fog of war

VICTORY:
  Domination: eliminate all enemies
  Science: research all technologies
  200 turn limit

Press ? or Esc to close help
`

const zhHelpText = `
CIV-TUI 帮助
=============

移动:
  方向键 / hjkl - 移动光标 / 选中单位
  绿色地块显示可到达位置

单位操作:
  F - 建立城市（仅开拓者）
  W - 等待/跳过单位回合
  N - 选择下一个可移动单位
  R - 远程攻击模式（弓箭手）
  I - 建造设施（工人）

城市操作:
  B - 打开建造菜单（在城市上时）
  Enter（在己方城市上，无单位时）- 城市详情

查看:
  V - 查看地块详情

科研:
  T - 打开科技研究菜单

外交:
  D - 打开外交菜单

游戏:
  S - 保存游戏
  Enter - 结束回合

显示:
  蓝色单位/城市 = 己方
  红色单位/城市 = 敌方
  绿色高亮 = 可到达地块
  红色高亮 = 远程攻击范围
  暗淡 = 已探索但不可见
  深色地块 = 战争迷雾

胜利条件:
  统治: 消灭所有敌人
  科技: 研究所有科技
  200 回合限制

按 ? 或 Esc 关闭帮助
`

var zhCN = map[string]string{
	// --- Main menu ---
	"A Terminal Civilization Game": "终端文明游戏",
	"New Game":                     "新游戏",
	"Load Game":                    "加载存档",
	"Settings":                     "设置",
	"Quit":                         "退出",
	"[↑/↓] Navigate  [Enter] Select  [Q] Quit": "[↑/↓] 导航  [Enter] 选择  [Q] 退出",
	"Loading...":      "加载中...",
	"Loading game...": "加载游戏中...",

	// --- Settings ---
	"SETTINGS":       "设置",
	"Language: %s":   "语言: %s",
	"Map Size: %s":   "地图大小: %s",
	"AI Civs: %d":    "AI 文明: %d",
	"Difficulty: %s": "难度: %s",
	"Small":          "小",
	"Medium":         "中",
	"Large":          "大",
	"Easy":           "简单",
	"Normal":         "普通",
	"Hard":           "困难",
	"Back":           "返回",
	"[←/→] Change value  [Enter/Esc] Back": "[←/→] 调整  [Enter/Esc] 返回",

	// --- Header ---
	"Turn: %d  Gold: %d (+%d)  Sci: %d  %s":          "回合: %d  金币: %d (+%d)  科研: %d  %s",
	"  [RANGED MODE - Enter to fire, Esc to cancel]": "  [远程模式 - Enter 射击, Esc 取消]",

	// --- Info panel: selected unit ---
	"SELECTED UNIT":               "选中单位",
	"Move: %d/%d  XP: %d  Lv: %d": "移动: %d/%d  经验: %d  等级: %d",
	"Atk: %d  Def: %d":            "攻击: %d  防御: %d",
	"Pos: (%d, %d)":               "位置: (%d, %d)",
	"Terrain: %s":                 "地形: %s",
	"Defense bonus: +%d%%":        "防御加成: +%d%%",
	"Cursor: %s (cost %d)":        "光标: %s (消耗 %d)",
	"Cursor: (%d, %d)":            "光标: (%d, %d)",
	"Yields: F%d P%d G%d":         "产出: 粮%d 产%d 金%d",
	"Unexplored":                  "未探索",
	"City: %s":                    "城市: %s",
	"Pop: %d  HP: %d/%d":          "人口: %d  HP: %d/%d",

	// --- Info panel: actions ---
	"ACTIONS":               "操作",
	"[F] Found City":        "[F] 建立城市",
	"[I] Build Improvement": "[I] 建造设施",
	"[R] Ranged Attack":     "[R] 远程攻击",
	"[W] Wait/Skip":         "[W] 等待/跳过",
	"[N] Next Unit":         "[N] 下一单位",
	"[B] Build Menu":        "[B] 建造菜单",
	"[T] Tech Menu":         "[T] 科技菜单",
	"[D] Diplomacy":         "[D] 外交",
	"[S] Save Game":         "[S] 保存游戏",
	"[V] Inspect Tile":      "[V] 查看地块",
	"[Enter] End Turn":      "[Enter] 结束回合",
	"[?] Help":              "[?] 帮助",

	// --- Info panel: research ---
	"RESEARCH":              "科研",
	"Researching: %s":       "研究中: %s",
	"Progress: %d/%d":       "进度: %d/%d",
	"Press [T] to research": "按 [T] 开始研究",
	"Techs: %d/%d":          "科技: %d/%d",

	// --- Minimap ---
	"MINIMAP": "小地图",

	// --- Messages ---
	"Messages:": "消息:",

	// --- Build menu ---
	"BUILD MENU":                   "建造菜单",
	"%s (cost %d)":                 "%s (花费 %d)",
	"[Enter]=select  [Esc]=cancel": "[Enter]=选择  [Esc]=取消",

	// --- Tech menu ---
	"TECH MENU":          "科技菜单",
	"No techs available": "无可用科技",

	// --- City details ---
	"CITY: %s":                                  "城市: %s",
	"Population: %d  HP: %d/%d":                 "人口: %d  HP: %d/%d",
	"Food: %d/%d  Production: %d":               "粮食: %d/%d  产能: %d",
	"Yields: Food %d  Prod %d  Gold %d  Sci %d": "产出: 粮食 %d  产能 %d  金币 %d  科研 %d",
	"Buildings:":                                "建筑:",
	"(none)":                                    "(无)",
	"Production Queue:":                         "生产队列:",
	"(empty)":                                   "(空)",
	"[Enter/Esc] Close":                         "[Enter/Esc] 关闭",

	// --- Diplomacy ---
	"DIPLOMACY":                             "外交",
	"No other civilizations":                "没有其他文明",
	"Peace":                                 "和平",
	"War":                                   "战争",
	" [Enter=declare war]":                  " [Enter=宣战]",
	" [Enter=make peace]":                   " [Enter=议和]",
	"[Enter]=toggle war/peace  [Esc]=close": "[Enter]=切换战争/和平  [Esc]=关闭",

	// --- Game over ---
	"VICTORY! Press Q to quit.":                   "胜利！按 Q 退出。",
	"DEFEAT! Press Q to quit.":                    "失败！按 Q 退出。",
	"DRAW - Turn limit reached! Press Q to quit.": "平局 - 达到回合上限！按 Q 退出。",

	// --- Tile inspect ---
	"TILE INSPECT":                            "地块详情",
	"Position: (%d, %d)":                      "位置: (%d, %d)",
	"Unexplored territory":                    "未探索区域",
	"[Esc/V] Close":                           "[Esc/V] 关闭",
	"Terrain":                                 "地形",
	"  Food: %d  Prod: %d  Gold: %d":          "  粮食: %d  产能: %d  金币: %d",
	"  Move Cost: %d":                         "  移动消耗: %d",
	"  Defense: +%d%%":                        "  防御: +%d%%",
	"Impassable":                              "不可通行",
	"Improvement":                             "设施",
	"  Food +%d  Prod +%d  Gold +%d":          "  粮食 +%d  产能 +%d  金币 +%d",
	"Unit":                                    "单位",
	"  HP: %d/%d  Move: %d/%d":                "  HP: %d/%d  移动: %d/%d",
	"  Atk: %d  Def: %d  XP: %d  Lv: %d":      "  攻: %d  防: %d  经验: %d  等级: %d",
	"  Range: %d":                             "  射程: %d",
	"  Building: %s (%d turns left)":          "  建造中: %s (剩余 %d 回合)",
	"City":                                    "城市",
	"  Pop: %d  HP: %d/%d  Def: %d":           "  人口: %d  HP: %d/%d  防御: %d",
	"  Food: %d  Prod: %d  Gold: %d  Sci: %d": "  粮食: %d  产能: %d  金币: %d  科研: %d",
	"  Buildings: ":                           "  建筑: ",
	"Visible":                                 "可见",
	"Revealed (not in sight)":                 "已探索（不在视野内）",
	"Visibility: ":                            "可见性: ",

	// --- Action messages (update.go) ---
	"This unit cannot perform ranged attacks":                                  "该单位无法进行远程攻击",
	"Ranged mode: select target with arrow keys, Enter to fire, Esc to cancel": "远程模式: 方向键选择目标, Enter 射击, Esc 取消",
	"Target out of range":           "目标超出射程",
	"No enemy unit at target":       "目标位置没有敌方单位",
	"Made peace with %s":            "与 %s 达成和平",
	"Declared war on %s":            "向 %s 宣战",
	"%s promoted!":                  "%s 晋升了！",
	"Failed to save: %s":            "保存失败: %s",
	"Game saved!":                   "游戏已保存！",
	"Need %s to build %s":           "需要 %s 才能建造 %s",
	"Worker building %s (%d turns)": "工人建造 %s (%d 回合)",
	"Queued: %s in %s":              "排队生产: %s (%s)",

	// --- Game messages (game.go) ---
	"Can't move there":                                   "无法移动到那里",
	"Terrain not passable":                               "地形不可通行",
	"Not enough movement":                                "移动力不足",
	"Not at war":                                         "未处于战争状态",
	"Tile occupied by friendly unit":                     "格子被友方单位占据",
	"%s attacks %s → killed!":                            "%s 攻击 %s → 击杀！",
	"%s attacks %s → attacker killed!":                   "%s 攻击 %s → 攻击者阵亡！",
	"%s attacks %s → both damaged":                       "%s 攻击 %s → 双方受伤",
	"%s ranged attacks %s → killed!":                     "%s 远程攻击 %s → 击杀！",
	"%s ranged attacks %s → hit!":                        "%s 远程攻击 %s → 命中！",
	"Captured %s!":                                       "占领了 %s！",
	"Attacked %s":                                        "进攻了 %s",
	"You have been defeated!":                            "你被击败了！",
	"Domination Victory! You conquered all enemies!":     "统治胜利！你征服了所有敌人！",
	"Science Victory! You researched all technologies!":  "科技胜利！你研究了所有科技！",
	"Turn limit reached! Game over.":                     "达到回合上限！游戏结束。",
	"%s trained %s":                                      "%s 训练了 %s",
	"%s built %s":                                        "%s 建造了 %s",
	"%s discovered %s!":                                  "%s 发现了 %s！",
	"Worker built %s at (%d,%d)":                         "工人建造了 %s (%d,%d)",
	"Turn 1: Welcome to Civ-TUI! Found a city with [F].": "回合 1: 欢迎来到 Civ-TUI！按 [F] 建立城市。",
	"Only Settlers can found cities":                     "只有开拓者才能建立城市",
	"Cannot found city here":                             "无法在此建立城市",
	"Too close to another city":                          "距离其他城市太近",
	"Founded %s!":                                        "建立了 %s！",
	"City ":                                              "城市 ",
	"%s grew to population %d":                           "%s 人口增长到 %d",

	// --- Entity names: Terrains ---
	"Ocean":     "海洋",
	"Coast":     "海岸",
	"Grassland": "草地",
	"Plains":    "平原",
	"Hills":     "丘陵",
	"Mountains": "山脉",
	"Forest":    "森林",
	"Desert":    "沙漠",
	"Tundra":    "冻原",

	// --- Entity names: Units ---
	"Settler":   "开拓者",
	"Scout":     "斥候",
	"Warrior":   "战士",
	"Archer":    "弓箭手",
	"Spearman":  "长矛兵",
	"Swordsman": "剑士",
	"Horseman":  "骑兵",
	"Worker":    "工人",

	// --- Entity names: Buildings ---
	"Granary":  "谷仓",
	"Barracks": "兵营",
	"Market":   "市场",
	"Library":  "图书馆",
	"Walls":    "城墙",

	// --- Entity names: Techs ---
	"Agriculture":      "农业",
	"Pottery":          "陶器",
	"Mining":           "采矿",
	"Bronze Working":   "青铜冶炼",
	"Iron Working":     "铁器冶炼",
	"Writing":          "文字",
	"Archery":          "箭术",
	"Animal Husbandry": "畜牧",
	"Horseback Riding": "骑术",
	"Calendar":         "历法",
	"The Wheel":        "轮子",

	// --- Entity names: Improvements ---
	"Farm":        "农场",
	"Mine":        "矿场",
	"Road":        "道路",
	"Lumber Mill": "伐木场",

	// --- Entity names: Civilizations ---
	"Roman Empire": "罗马帝国",
	"Mongols":      "蒙古",
	"Egypt":        "埃及",
	"China":        "中国",
	"Greece":       "希腊",
}
