import { storiesOf } from '@storybook/react'

type StoryCategory = 'web' | 'shared' | 'browser' | 'branded'

export const addStory = (category: StoryCategory, kind: string, module: NodeModule): ReturnType<typeof storiesOf> => {
    return storiesOf(`${category}/${kind}`, module)
}
