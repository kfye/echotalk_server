package course

import "time"

// dayStart 把时间截到当日 0 点（按其所在时区），用于按自然日计算解锁天数。
func dayStart(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// unlockedDay 按开营日算今天已解锁到第几天：开营当天=第1天，每过一自然日+1，封顶 totalDays。
// 开营日在未来（还没开营）返回 0。
func unlockedDay(startDate time.Time, totalDays int, now time.Time) int {
	s := dayStart(startDate)
	n := dayStart(now)
	if n.Before(s) {
		return 0
	}
	days := int(n.Sub(s).Hours()/24) + 1
	if days > totalDays {
		days = totalDays
	}
	return days
}

// lessonUnlocked 判断第 day 天的课在当前进度下是否已解锁。
func lessonUnlocked(day, unlocked int) bool {
	return day >= 1 && day <= unlocked
}
