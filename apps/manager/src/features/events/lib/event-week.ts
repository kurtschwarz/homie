import type { ManagerEvent } from '../types'

/**
 * The seven calendar days starting with the day containing `now`, each at
 * local midnight. Days are stepped with `setDate` rather than by adding 24
 * hours, so a daylight-saving change cannot skip or repeat a date.
 *
 * @param now - The moment that counts as today. Defaults to the current time;
 * pass a fixed value to get a stable result.
 */
export function weekDays(now = new Date()) {
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  return Array.from({ length: 7 }, (_, offset) => {
    const day = new Date(start)
    day.setDate(start.getDate() + offset)
    return day
  })
}

/**
 * The events attributed to one person. An event shared by several people is
 * included for each of them.
 *
 * @param events - Events as returned by the events query.
 * @param personId - The id of the person to keep events for.
 */
export function eventsFor(events: ManagerEvent[], personId: string) {
  return events.filter((event) => event.people.some((person) => person.id === personId))
}

/**
 * The events that start on a given local calendar day. Only the start counts:
 * an event that runs past midnight or spans several days appears on the day
 * it begins and not on the days after.
 *
 * @param events - Events to filter.
 * @param day - Any moment within the day to match; the time of day is ignored.
 */
export function onDay(events: ManagerEvent[], day: Date) {
  return events.filter(
    (event) =>
      event.start.getFullYear() === day.getFullYear() &&
      event.start.getMonth() === day.getMonth() &&
      event.start.getDate() === day.getDate(),
  )
}
