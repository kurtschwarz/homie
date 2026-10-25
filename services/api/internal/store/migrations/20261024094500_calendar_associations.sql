-- +goose Up
CREATE TABLE calendar_associations (
  calendar_id BLOB NOT NULL REFERENCES calendars (id) ON DELETE CASCADE,
  person_id   BLOB NOT NULL REFERENCES people (id) ON DELETE CASCADE,
  scope       TEXT NOT NULL CHECK (scope IN ('all', 'matched')),
  PRIMARY KEY (calendar_id, person_id)
) STRICT;

-- kind is part of the key so the same text can be both a tag to match and a
-- word to look for in a title.
CREATE TABLE calendar_association_terms (
  calendar_id BLOB NOT NULL,
  person_id   BLOB NOT NULL,
  kind        TEXT NOT NULL CHECK (kind IN ('tag', 'word')),
  term        TEXT NOT NULL,
  label       TEXT NOT NULL,
  PRIMARY KEY (calendar_id, person_id, kind, term),
  FOREIGN KEY (calendar_id, person_id)
    REFERENCES calendar_associations (calendar_id, person_id) ON DELETE CASCADE
) STRICT;

INSERT INTO calendar_associations (calendar_id, person_id, scope)
SELECT DISTINCT calendar_id, person_id, 'matched' FROM tag_mappings WHERE kind = 'person';

INSERT INTO calendar_association_terms (calendar_id, person_id, kind, term, label)
SELECT calendar_id, person_id, 'tag', tag, label FROM tag_mappings WHERE kind = 'person';

-- An include mapping showed events under the family badge when nobody else
-- matched, so it becomes an association with whoever is the family person.
INSERT OR IGNORE INTO calendar_associations (calendar_id, person_id, scope)
SELECT DISTINCT mapping.calendar_id, person.id, 'matched'
FROM tag_mappings AS mapping
JOIN people AS person ON person.is_family = 1
WHERE mapping.kind = 'include';

INSERT OR IGNORE INTO calendar_association_terms (calendar_id, person_id, kind, term, label)
SELECT mapping.calendar_id, person.id, 'tag', mapping.tag, mapping.label
FROM tag_mappings AS mapping
JOIN people AS person ON person.is_family = 1
WHERE mapping.kind = 'include';

DROP TABLE tag_mappings;

-- +goose Down
-- This rollback is intentionally lossy: tag_mappings cannot represent word
-- terms, whole-calendar associations, or two people sharing one tag.
CREATE TABLE tag_mappings (
  calendar_id BLOB NOT NULL REFERENCES calendars (id) ON DELETE CASCADE,
  tag         TEXT NOT NULL,
  label       TEXT NOT NULL,
  kind        TEXT NOT NULL CHECK (kind IN ('person', 'include')),
  person_id   BLOB REFERENCES people (id) ON DELETE CASCADE,
  PRIMARY KEY (calendar_id, tag),
  CHECK ((kind = 'person') = (person_id IS NOT NULL))
) STRICT;

INSERT INTO tag_mappings (calendar_id, tag, label, kind, person_id)
SELECT calendar_id, term, label, 'person', person_id
FROM calendar_association_terms
WHERE kind = 'tag'
GROUP BY calendar_id, term;

DROP TABLE calendar_association_terms;
DROP TABLE calendar_associations;
