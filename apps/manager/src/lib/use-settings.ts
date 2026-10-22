import type { Swatch } from "@homie/ui/lib/people";
import { useCallback, useEffect, useState } from "react";
import { graphql } from "./graphql";

export type ManagerPerson = {
  id: string;
  name: string;
  emoji: string;
  swatch: Swatch;
  isFamily: boolean;
};

export type TagKind = "PERSON" | "INCLUDE";

export type TagMapping = {
  tag: string;
  label: string;
  kind: TagKind;
  person: ManagerPerson | null;
};

export type Settings = {
  people: ManagerPerson[];
  tagMappings: TagMapping[];
  calendarTags: string[];
};

type SettingsState =
  | { kind: "loading" }
  | { kind: "ready"; settings: Settings; saving: boolean; error?: string }
  | { kind: "error"; message: string };

const SETTINGS_QUERY = `
  query Settings {
    people { id name emoji swatch isFamily }
    tagMappings { tag label kind person { id name emoji swatch isFamily } }
    calendarTags
  }
`;

export type PersonInput = {
  id?: string;
  name: string;
  emoji: string;
  swatch: string;
  isFamily: boolean;
};
export type TagMappingInput = { tag: string; kind: TagKind; personId?: string };

/**
 * One hook for people and tag mappings: they are edited together (assigning
 * a tag to a person needs the people list to be current), and every
 * mutation refetches both. Refetching rather than patching local state is
 * the right call for a single-user settings page where correctness beats a
 * frame of latency.
 */
export function useSettings() {
  const [state, setState] = useState<SettingsState>({ kind: "loading" });

  const load = useCallback(async () => {
    try {
      const data = await graphql<Settings>(SETTINGS_QUERY);
      setState({ kind: "ready", settings: data, saving: false });
    } catch (err) {
      setState({ kind: "error", message: err instanceof Error ? err.message : String(err) });
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const mutate = useCallback(
    async <T>(query: string, variables: Record<string, unknown>): Promise<T | undefined> => {
      setState((prev) =>
        prev.kind === "ready" ? { ...prev, saving: true, error: undefined } : prev,
      );
      try {
        const data = await graphql<T>(query, variables);
        await load();
        return data;
      } catch (err) {
        const message = err instanceof Error ? err.message : String(err);
        setState((prev) =>
          prev.kind === "ready"
            ? { ...prev, saving: false, error: message }
            : { kind: "error", message },
        );
        return undefined;
      }
    },
    [load],
  );

  const savePerson = useCallback(
    (input: PersonInput) =>
      mutate(`mutation SavePerson($input: PersonInput!) { savePerson(input: $input) { id } }`, {
        input,
      }),
    [mutate],
  );

  const deletePerson = useCallback(
    (id: string) => mutate(`mutation DeletePerson($id: ID!) { deletePerson(id: $id) }`, { id }),
    [mutate],
  );

  const saveTagMapping = useCallback(
    (input: TagMappingInput) =>
      mutate(
        `mutation SaveTagMapping($input: TagMappingInput!) { saveTagMapping(input: $input) { tag } }`,
        { input },
      ),
    [mutate],
  );

  const deleteTagMapping = useCallback(
    (tag: string) =>
      mutate(`mutation DeleteTagMapping($tag: String!) { deleteTagMapping(tag: $tag) }`, { tag }),
    [mutate],
  );

  // Going through mutate means the refetch afterwards picks up any tag the
  // sync just discovered, which is the whole point of refreshing from here.
  const refreshEvents = useCallback(
    () => mutate(`mutation RefreshEvents { refreshEvents }`, {}),
    [mutate],
  );

  return { state, savePerson, deletePerson, saveTagMapping, deleteTagMapping, refreshEvents };
}
