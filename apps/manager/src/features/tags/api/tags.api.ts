import { graphql } from '@/shared/lib/graphql'
import type { TagMapping, TagMappingInput } from '../types'

const TAG_MAPPINGS_QUERY = `
  query TagMappings {
    tagMappings { tag label kind person { id name emoji swatch isFamily } }
  }
`

export const tagsApi = {
  list: async (): Promise<TagMapping[]> => {
    const data = await graphql<{ tagMappings: TagMapping[] }>(TAG_MAPPINGS_QUERY)
    return data.tagMappings
  },

  save: (input: TagMappingInput) =>
    graphql<{ saveTagMapping: { tag: string } }>(
      `mutation SaveTagMapping($input: TagMappingInput!) { saveTagMapping(input: $input) { tag } }`,
      { input },
    ),

  remove: (tag: string) =>
    graphql(`mutation DeleteTagMapping($tag: String!) { deleteTagMapping(tag: $tag) }`, { tag }),
}
