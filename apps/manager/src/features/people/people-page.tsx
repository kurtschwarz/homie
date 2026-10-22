import { Button } from "@homie/ui/components/button";
import { useAppData } from "../../app-data";
import { SettingsGate } from "../../components/settings-gate";
import type { ManagerPerson, PersonInput } from "../../lib/use-settings";
import { Page } from "../../shell/page";
import { PersonRow } from "./person-row";

export function PeoplePage() {
  const { savePerson, deletePerson } = useAppData();
  const actions = (
    <Button
      variant="solid"
      onClick={() =>
        document
          .getElementById("new-person")
          ?.scrollIntoView({ behavior: "smooth", block: "center" })
      }
    >
      + New person
    </Button>
  );

  return (
    <Page
      title="People"
      description="Manage your household and the people shown on Homie."
      actions={actions}
    >
      <SettingsGate>
        {(settings) => (
          <PeopleSection people={settings.people} onSave={savePerson} onDelete={deletePerson} />
        )}
      </SettingsGate>
    </Page>
  );
}

function PeopleSection({
  people,
  onSave,
  onDelete,
}: {
  people: ManagerPerson[];
  onSave: (input: PersonInput) => void;
  onDelete: (id: string) => void;
}) {
  return (
    <section className="flex flex-col gap-4">
      {people.length === 0 && (
        <p className="m-0 text-hm-body text-hm-ink-2">
          No one is set up yet. Add the first person below and mark them as family — events tagged{" "}
          <strong>Homie</strong> attribute to whoever holds that flag.
        </p>
      )}
      <div className="flex flex-col gap-3">
        {people.map((person) => (
          <PersonRow
            key={person.id}
            person={person}
            familyChecked={person.isFamily}
            onSave={onSave}
            onSetFamily={() =>
              onSave({
                id: person.id,
                name: person.name,
                emoji: person.emoji,
                swatch: person.swatch,
                isFamily: true,
              })
            }
            onDelete={() => onDelete(person.id)}
          />
        ))}
        <div id="new-person">
          <PersonRow onSave={onSave} />
        </div>
      </div>
    </section>
  );
}
