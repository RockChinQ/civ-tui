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
