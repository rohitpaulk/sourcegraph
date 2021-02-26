import { Tab, TabList, TabPanel, TabPanels, Tabs } from '@reach/tabs'
import * as H from 'history'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useHistory, useLocation } from 'react-router'
import { Button } from 'reactstrap'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol/contribution'
import { ActivationProps } from '../../../../shared/src/components/activation/Activation'
import { FetchFileParameters } from '../../../../shared/src/components/CodeExcerpt'
import { Resizable } from '../../../../shared/src/components/Resizable'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { VersionContextProps } from '../../../../shared/src/search/util'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { ThemeProps } from '../../../../shared/src/theme'
import { useObservable } from '../../../../shared/src/util/useObservable'
import { Tab as Tab1 } from '../Tabs'
import { EmptyPanelView } from './views/EmptyPanelView'
import { PanelView } from './views/PanelView'

interface Props
    extends ExtensionsControllerProps,
        PlatformContextProps,
        SettingsCascadeProps,
        ActivationProps,
        TelemetryProps,
        ThemeProps,
        VersionContextProps {
    location: H.Location
    history: H.History
    repoName?: string
    fetchHighlightedFileLineRanges: (parameters: FetchFileParameters, force?: boolean) => Observable<string[][]>
}

/**
 * A tab and corresponding content to display in the panel.
 */
interface PanelItem extends Tab1<string> {
    /**
     * Controls the relative order of panel items. The items are laid out from highest priority (at the beginning)
     * to lowest priority (at the end). The default is 0.
     */
    priority: number

    /** The content element to display when the tab is active. */
    element: JSX.Element

    /**
     * Whether this panel contains a list of locations (from a location provider). This value is
     * exposed to contributions as `panel.activeView.hasLocations`. It is true if there is a
     * location provider (even if the result set is empty).
     */
    hasLocations?: boolean
}

/**
 * The panel, which is a tabbed component with contextual information. Components rendering the panel should
 * generally use ResizablePanel, not Panel.
 *
 * Other components can contribute panel items to the panel.
 */

const Panel: React.FunctionComponent<Props> = props => {
    const [panels, setPanels] = useState<PanelItem[]>([])
    const [tabIndex, setTabIndex] = useState(0)
    const { hash, pathname } = useLocation()
    const history = useHistory()
    const handlePanelClose = useCallback(() => history.replace(pathname), [history, pathname])

    const items = useObservable(
        useMemo(
            () =>
                props.extensionsController.services.panelViews
                    .getPanelViews(ContributableViewContainer.Panel)
                    .pipe(map(panelViews => ({ panelViews }))),
            [props.extensionsController.services.panelViews]
        )
    )

    const handleActiveTab = useCallback(
        (index: number): void => {
            history.replace(`${pathname}${hash.split('=')[0]}=${panels[index].id}`)
        },
        [hash, history, panels, pathname]
    )

    useEffect(() => {
        setTabIndex(panels.findIndex(({ id }) => id === `${hash.split('=')[1]}`))
    }, [hash, panels])

    useEffect(() => {
        if (items?.panelViews) {
            setPanels(
                items.panelViews
                    .map(
                        (panelView): PanelItem => ({
                            label: panelView.title,
                            id: panelView.id,
                            priority: panelView.priority,
                            element: <PanelView {...props} panelView={panelView} />,
                            hasLocations: !!panelView.locationProvider,
                        })
                    )
                    .sort((a, b) => b.priority - a.priority)
            )
        }
    }, [items?.panelViews, props])

    if (!items) {
        return <EmptyPanelView />
    }

    return (
        <Tabs className="w-100 overflow-hidden" index={tabIndex} onChange={handleActiveTab}>
            <div className="d-flex">
                <TabList>
                    {panels.map(({ label, id }) => (
                        <Tab key={id}>{label}</Tab>
                    ))}
                </TabList>
                <Button
                    onClick={handlePanelClose}
                    close={true}
                    className="bg-transparent border-0 close ml-auto"
                    title="Close sidebar (Alt+S/Opt+S)"
                />
            </div>
            <TabPanels className="h-100 overflow-auto">
                {panels.map(({ id, element }) => (
                    <TabPanel key={id}>{element}</TabPanel>
                ))}
            </TabPanels>
        </Tabs>
    )
}

/** A wrapper around Panel that makes it resizable. */
export const ResizablePanel: React.FunctionComponent<Props> = props => (
    <div className="w-100 bg-code">
        <Resizable position="top" defaultSize={350}>
            <Panel {...props} />
        </Resizable>
    </div>
)
