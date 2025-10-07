/**
 * Alerts management hook that encapsulates data fetching and mutations
 * for alert rules and events, extending useAlertsWithNotifications
 */

import { useCallback } from "react";
import { useAlertsWithNotifications } from "./useAlertsWithNotifications";
import { createAlertRule, updateAlertRule, deleteAlertRule, ackAlertEvent, type CreateAlertRuleRequest } from "@/lib/api";

export function useAlertsManagement(eventsLimit?: number) {
  const alertsHook = useAlertsWithNotifications(eventsLimit);

  // Rule management methods
  const createRule = useCallback(async (ruleData: CreateAlertRuleRequest) => {
    const newRule = await createAlertRule(ruleData);
    // Refresh data to show the new rule
    await alertsHook.refresh();
    return newRule;
  }, [alertsHook]);

  const updateRule = useCallback(async (id: number, ruleData: CreateAlertRuleRequest) => {
    const updatedRule = await updateAlertRule(id, ruleData);
    // Refresh data to show the updated rule
    await alertsHook.refresh();
    return updatedRule;
  }, [alertsHook]);

  const deleteRule = useCallback(async (id: number) => {
    await deleteAlertRule(id);
    // Refresh data to remove the deleted rule
    await alertsHook.refresh();
  }, [alertsHook]);

  // Event management methods
  const acknowledgeEvent = useCallback(async (id: number) => {
    await ackAlertEvent(id);
    // Refresh data to update the acknowledged event
    await alertsHook.refresh();
  }, [alertsHook]);

  return {
    // Expose all the original hook properties and methods
    ...alertsHook,
    
    // Add the management methods
    createRule,
    updateRule,
    deleteRule,
    acknowledgeEvent,
  };
}