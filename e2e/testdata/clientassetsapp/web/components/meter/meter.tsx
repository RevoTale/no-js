/** @jsx jsx */

type JSXMarker = {
	tag: string;
	props: Record<string, string>;
};

const jsx = (tag: string, props: Record<string, string>): JSXMarker => ({ tag, props });
const marker = <span data-client-assets-meter="loaded" />;

document.documentElement.dataset.clientAssetsMeter = marker.props["data-client-assets-meter"];
