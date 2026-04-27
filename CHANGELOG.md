# Changelog

## [2.0.0](https://github.com/RevoTale/no-js/compare/v1.3.0...v2.0.0) (2026-04-27)


### ⚠ BREAKING CHANGES

* **appshape:** apps must move loose `web/components` files into `web/components/<name>/`, keep component assets same-stem with the package anchor, and move public component Go API into `<name>.go`.

### Features

* **app:** enforce strict app shape and route-static assets ([79dd63c](https://github.com/RevoTale/no-js/commit/79dd63cde1b012edee83a899055fa1694122ea86))
* **app:** enforce strict shape and route-static assets ([b6d72b6](https://github.com/RevoTale/no-js/commit/b6d72b6ca9f278b5882cba2cb81094663bbb4580))
* **appshape:** enforce strict app input shape ([8ed59cc](https://github.com/RevoTale/no-js/commit/8ed59ccc30daa9b5d82c5759afdf16cf80c40ad5))
* **assets:** default templ CSS extraction from config ([557b272](https://github.com/RevoTale/no-js/commit/557b27291785a231904e91f6808a4153fcadcf60))
* **client-assets:** add route-static CSS and script bundles ([3d460e7](https://github.com/RevoTale/no-js/commit/3d460e7020cc57d9024c4ef6796e99993d2437f9))
* **client-assets:** fold CSS into layout subtree bundles ([deb57cd](https://github.com/RevoTale/no-js/commit/deb57cd1b0b568c7272f2b2a855ffc8a864a06e2))


### Bug Fixes

* **appshape:** allow slot fallback client assets ([98b330b](https://github.com/RevoTale/no-js/commit/98b330b82f06fbb1ccb59d9b96b2f8e6b11b6b0f))
* **appshape:** reject invalid asset-only route shapes ([55a8ed2](https://github.com/RevoTale/no-js/commit/55a8ed2642835854ef7b9d48348dea169e023d67))
* **client-assets:** include all slot assets for owning layouts ([4b30fe9](https://github.com/RevoTale/no-js/commit/4b30fe93954efe16f0f36ef37915458b47c21877))
* **client-assets:** parse selectors before anonymizing classes ([8fb862d](https://github.com/RevoTale/no-js/commit/8fb862d9fd5d6702f235d5750413bd2f2fc2fd1a))
* intoriduct a new mental model for the css composition ([de24750](https://github.com/RevoTale/no-js/commit/de24750c7d0d006109a7fabda83721fe6e7bce54))

## [1.3.0](https://github.com/RevoTale/no-js/compare/v1.2.0...v1.3.0) (2026-04-21)


### Features

* **templ:** auto-wire global css from the generated bundle ([44b3622](https://github.com/RevoTale/no-js/commit/44b3622635583c9e9aa0d1b05fe220564527eef7))
* **templ:** auto-wire global css from the generated bundle ([7a20690](https://github.com/RevoTale/no-js/commit/7a206904f0bbec53fd1f6b31742092524c2591a6))

## [1.2.0](https://github.com/RevoTale/no-js/compare/v1.1.0...v1.2.0) (2026-04-21)


### Features

* **templ:** generate route CSS in web/generated ([3534ab7](https://github.com/RevoTale/no-js/commit/3534ab7a5726cc9f4b6f8eebd06b3d70157bb69b))

## [1.1.0](https://github.com/RevoTale/no-js/compare/v1.0.0...v1.1.0) (2026-04-20)


### Features

* **templ:** bundle component CSS into hashed assets ([aa7259c](https://github.com/RevoTale/no-js/commit/aa7259cf22c8d55771274835093ee47f9a895131))
* **templ:** bundle global CSS into hashed assets and add real fixture coverage ([c673bfb](https://github.com/RevoTale/no-js/commit/c673bfb81e2e8a3285a779a8d3922521fd8eb214))

## [1.0.0](https://github.com/RevoTale/no-js/compare/v0.1.0...v1.0.0) (2026-04-13)


### ⚠ BREAKING CHANGES

* no-js now defaults to web/* paths, build-time packages moved under internal/, generated server bootstrap was removed, and blog packages moved from internal/web/ to web/*.

### Features

* `no-js` support for the neste fedd and sitemaps. Fix the bug with wrong routes priorirty ([e9b53a2](https://github.com/RevoTale/no-js/commit/e9b53a2989e3e1b4f712c0c2adde4ba3c06babf6))
* `public` directory with static asssets that serves from the root and replacated what `public` dir does in nextjs ([25d36ff](https://github.com/RevoTale/no-js/commit/25d36ff3665cab6d7cbb7038c0dcf3952fafc246))
* 404 page ([5db4662](https://github.com/RevoTale/no-js/commit/5db4662b20073cb57cadbb392c6602d4d54c3465))
* add a sidebar filters with the tags, short/long tales. Add canonical pagess for all them ([ec3a535](https://github.com/RevoTale/no-js/commit/ec3a5354a81673f2ff74fb2a2970d7ecbfcd56e2))
* add public code link ([ff27927](https://github.com/RevoTale/no-js/commit/ff279278aca4b239a1b4e97daf7b521ac4ddc5a0))
* add the search bar ([51f8884](https://github.com/RevoTale/no-js/commit/51f8884e9f2eebbfe45a4f22d88e0cdcd8806dea))
* adopt strict web/* app layout across blog and no-js ([0213489](https://github.com/RevoTale/no-js/commit/0213489a3f1c292f3e89392ba2b7fc2d3900dbcf))
* auttocomplete for templ ([eb96b94](https://github.com/RevoTale/no-js/commit/eb96b942ad1c2eb616c6a83c964e42a1ec7e4f58))
* badge for not visited notes. ([d2d2619](https://github.com/RevoTale/no-js/commit/d2d2619f6560016bcf86112fdde24ed740534fe3))
* CI ([c5a7a03](https://github.com/RevoTale/no-js/commit/c5a7a0393da7a21e1bf43d5292cc9a35e2390eac))
* clear button for the search bar ([fdc04b8](https://github.com/RevoTale/no-js/commit/fdc04b827bcf270ba774e37bb4c45452792c7bd1))
* configure privacy-first lovely eye analytics ([e36c5ac](https://github.com/RevoTale/no-js/commit/e36c5ac95aa1dbe326564c6ad5fadd79f2a1400c))
* debug lovely eye config ([71d3a12](https://github.com/RevoTale/no-js/commit/71d3a128bf190006f3745869a3981ac6a76fcdb1))
* deduplicate data loading via NextJs-like per-rendering cache which deduplicates external api requesta and expensive work. ([9019852](https://github.com/RevoTale/no-js/commit/90198524a28c883cf2b70606a0e85550234775c9))
* design the i18 core. Translate all pages the same way I do current on the NextJS-written blog. ([b01a437](https://github.com/RevoTale/no-js/commit/b01a437cbab93bd78d7d95395220ba187eba0cfe))
* devcontainer ([b32bf30](https://github.com/RevoTale/no-js/commit/b32bf300ded389cf5f83ca9c972f11e2c9533c90))
* display the badge  for unvisited pages on the bottom level and fix the issue with the browser privacy forbidding it. ([bbb80c2](https://github.com/RevoTale/no-js/commit/bbb80c2c5ed5bf7b8ed584aa3f08157d091dbe8e))
* display the real techonolgy stack used for the blog in the footer ([0e501b7](https://github.com/RevoTale/no-js/commit/0e501b72274336c5ed237cc857311422de0c8116))
* do not highligh logo button due to contrast ([b74d267](https://github.com/RevoTale/no-js/commit/b74d267daea739d7d47a11ad866d0a3546d33c57))
* document the architecture choices to not forget them later ([f847513](https://github.com/RevoTale/no-js/commit/f8475136e0c06e36101886f50237854953e93bd1))
* extract the http server to the `framework` module to do conver tto the  complete frameowkr later. ([7bac77b](https://github.com/RevoTale/no-js/commit/7bac77b26d6e7bd782f443fbe01924d79f65d5b8))
* fix CI ([3a6bf54](https://github.com/RevoTale/no-js/commit/3a6bf5412daacab3146b4fc6cb862bfe921d9cea))
* fix the micro note content displayed the seo fallbacke ([4bc0067](https://github.com/RevoTale/no-js/commit/4bc006796a458d4b02b0c53fec0935e2668d0384))
* fix the proxying the iamages to our infrastructure ([04475c5](https://github.com/RevoTale/no-js/commit/04475c54736b28f1615f50e4f30ddbc2e2e5e009))
* **framework:** add request-scoped metadata context for URL composition ([c598e41](https://github.com/RevoTale/no-js/commit/c598e41c03b5c935a0f3d3ac609c0e2126593c0f))
* gzip compression and static build info ([5ffcbe2](https://github.com/RevoTale/no-js/commit/5ffcbe2d9cfa2c30d6483d2c068ec7ac3e97456b))
* image lazy loading ([616c433](https://github.com/RevoTale/no-js/commit/616c4334b7e5f2c06973dbf50395a184ecde9e19))
* install deadcode validator and remove dead code ([7b22eb5](https://github.com/RevoTale/no-js/commit/7b22eb5be09cd03d747cd368b048b2a2bfab5841))
* language switcher and footer design adjustments ([b7f5dff](https://github.com/RevoTale/no-js/commit/b7f5dffb7e58d2529dde0b250a28a324c9c0de67))
* live state changes will live under`/.live` subpath to avoid the real routes and caching collisions. ([f24bab6](https://github.com/RevoTale/no-js/commit/f24bab6ff9fc4dcab2fe87f29e1c7e7479ae7e30))
* logging of the data resolvers timing. ([af6d30a](https://github.com/RevoTale/no-js/commit/af6d30a660149f3426f2750115f1948d88447a5b))
* make the author name always blue ([784081e](https://github.com/RevoTale/no-js/commit/784081ee9d5390b11bba2b91718145b9fff5a62f))
* make the NextJS-like metadata generator pattern and live replacement. ([7f6b9a8](https://github.com/RevoTale/no-js/commit/7f6b9a8600339cdd4dc7293cc8ea95b4e3560188))
* mark the generated code to avoid confison ([82e4e0f](https://github.com/RevoTale/no-js/commit/82e4e0fe3c695844c782aee8a241a452bcd53b15))
* migrate metadata of the layout and pages from the legacy RevoTale blog NextJS app. ([7f6b9a8](https://github.com/RevoTale/no-js/commit/7f6b9a8600339cdd4dc7293cc8ea95b4e3560188))
* migrate the copy button from internal blog ([9368bfb](https://github.com/RevoTale/no-js/commit/9368bfb247f5c4e8b90d54e8847766e8dafb7fad))
* proper devcontainer ([73d0ea5](https://github.com/RevoTale/no-js/commit/73d0ea5001582985659ad0a4bd94698295979a1c))
* reduce the amount of device sizes because we do not need so much ([16582a6](https://github.com/RevoTale/no-js/commit/16582a695bec3682c63f74b7d3359e15d970f480))
* refacotring to move more responsiblity, and simplify the config of the end app. Excracting the https://github.com/RevoTale/blog internals to the framwork core. Slow and hard process ([41fd2d8](https://github.com/RevoTale/no-js/commit/41fd2d8469a89e97b3a015adfa0ebd66809e4e1f))
* refactor the framwork to separate the runtime and bundling modules ([235abda](https://github.com/RevoTale/no-js/commit/235abda132b11290fc681a797b0ffcbd7f54ca02))
* remove the `datastar`. Migrate to the to the `htmx`. ([de33357](https://github.com/RevoTale/no-js/commit/de333578be22e523a3edacdc159bae6bace278c0))
* reogranize the datat resolvers to share the single namespace and one generate file with the definition. Much easier to read ([a4f3d0c](https://github.com/RevoTale/no-js/commit/a4f3d0cc3667cf23f3fbf4d8f7620268a833087c))
* Replace the `/static` url with the `/.revotale` url. ([7bac77b](https://github.com/RevoTale/no-js/commit/7bac77b26d6e7bd782f443fbe01924d79f65d5b8))
* **router:** add reserved app-router namespaces, slots, and route.go handlers ([139b1d5](https://github.com/RevoTale/no-js/commit/139b1d5130aa99b4e924f3b10805d6df391cdaa4))
* rss feed button ([f8cf21d](https://github.com/RevoTale/no-js/commit/f8cf21d35da52b9b54cf08be815320d5abe0121c))
* seo sitemaps and rss feeds. ([6c494dc](https://github.com/RevoTale/no-js/commit/6c494dc33f1402272a6369d110c75a28d0a3f749))
* separate image component similat to NextJS. Custom loader for the our internal infrastructure ([b700217](https://github.com/RevoTale/no-js/commit/b70021759f41422b0864ba70c54cd6e1475aaf57))
* setup dev environment separately from https://github.com/RevoTale/backend ([62c34e4](https://github.com/RevoTale/no-js/commit/62c34e4cbb48c8ceb134a660ae9ed4329ea0c4a3))
* ue the system level fonts for better readability and redcue the download size ([6b50b7f](https://github.com/RevoTale/no-js/commit/6b50b7f5e8d22c861a30f19524c8a51f98443f4e))
* use the esbuild for the statiuc files path hashing and minifiiing. ([edd11c9](https://github.com/RevoTale/no-js/commit/edd11c92a54c1ee8ed4c1fe2205097c82558f27c))


### Bug Fixes

* add all font variants: ([db2aa7e](https://github.com/RevoTale/no-js/commit/db2aa7e14486f8a34113d3db046e5c2e1e67b181))
* broken images due to dynamic preset size ([02b9247](https://github.com/RevoTale/no-js/commit/02b9247bb6b01197041e0f0f50aeb2a26d25d945))
* canonicalize note listing filter urls, place redirect to the canonical when suitable and add noindex for unknown params ([c8b70e0](https://github.com/RevoTale/no-js/commit/c8b70e0d12b547bca6e71688acf7eaa524838f84))
* **deps:** update all non-major dependencies ([87db0ef](https://github.com/RevoTale/no-js/commit/87db0ef8db8985638448c6cb354886f0116273cd))
* **deps:** update all non-major dependencies ([7479aa6](https://github.com/RevoTale/no-js/commit/7479aa69da5c66b66bf10e08155c56d6581850f2))
* **deps:** update all non-major dependencies ([79ac1b1](https://github.com/RevoTale/no-js/commit/79ac1b14661f6660c4f4d32aba9394da096119ab))
* **deps:** update all non-major dependencies ([a2dc121](https://github.com/RevoTale/no-js/commit/a2dc12197b730eac679cdf3437b1d9be37cfa4ea))
* **deps:** update module github.com/evanw/esbuild to v0.27.4 ([8548d86](https://github.com/RevoTale/no-js/commit/8548d86ed345080a3ccea1febbfc7bb33a1c2f17))
* **deps:** update module github.com/evanw/esbuild to v0.27.4 ([7c71525](https://github.com/RevoTale/no-js/commit/7c71525444ae152feacc442a677f00cd705ae27c))
* devcontainer image ([5b9e60a](https://github.com/RevoTale/no-js/commit/5b9e60a4ea09aa84680c3b35e128b739d3baad65))
* devcontainer setup ([a400e62](https://github.com/RevoTale/no-js/commit/a400e6255cf5f598123790c8be180eb51c03167c))
* **discovery:** stop generating sitemap IDs on unrelated requests ([842b600](https://github.com/RevoTale/no-js/commit/842b600bf1b74e88a3f8e580ebe9ab0877c11f0c))
* failing tests ([e0e3b72](https://github.com/RevoTale/no-js/commit/e0e3b7214f939dee750b0c43a9e9cf3f15698ec3))
* failing tests after reduce of the deviceSizes ([6e5e759](https://github.com/RevoTale/no-js/commit/6e5e75998c3df34b28b33ec47b00182b89dabdeb))
* golint validtion fails ([7bb9f23](https://github.com/RevoTale/no-js/commit/7bb9f23b39fe224f323594458590730e5bbaf98d))
* image sizes wrong ([aa7e35c](https://github.com/RevoTale/no-js/commit/aa7e35c1c82f5fe8610b7b376b2e4346e6daace3))
* indexing disabled due to robos tags ([37b7b3e](https://github.com/RevoTale/no-js/commit/37b7b3e410708ad55a219d496a08c0312f6b7719))
* L styles adjustment to be more mobile froiendly and remove unnesesarry elements ([db2aa7e](https://github.com/RevoTale/no-js/commit/db2aa7e14486f8a34113d3db046e5c2e1e67b181))
* remove the udnerline ([758bc80](https://github.com/RevoTale/no-js/commit/758bc80a312b67425ab178b4f84824db8dca8e97))
* **routes:** rename 404 template component to NotFound ([0b1fba1](https://github.com/RevoTale/no-js/commit/0b1fba1eb516537242b07196e504416389eb233d))
* search bar becomes block too late ([f9a9dcf](https://github.com/RevoTale/no-js/commit/f9a9dcfe928751cf464e19a0f3853447d5b67e89))
* second-level headers has tag h1 in the markdown rendering. That hurts SEO ([9f2def4](https://github.com/RevoTale/no-js/commit/9f2def4d84896e4d5c58f88f92df5c493a300ecd))
* single blog post  centered well ([f68657e](https://github.com/RevoTale/no-js/commit/f68657e3a5ad1870bc5e2cf0be2679a2e8548977))
* single page background and footer colors ([379619f](https://github.com/RevoTale/no-js/commit/379619fece8ba2893848be7c9a4918248a6a46cc))
* unrelated data has been loading sequently ([2b596d9](https://github.com/RevoTale/no-js/commit/2b596d9cfcd5f7cd3687bf9a51f6b70fa6ccd839))
* unrelated data in the notes list has been loading sequently ([9019852](https://github.com/RevoTale/no-js/commit/90198524a28c883cf2b70606a0e85550234775c9))
* wrong image sized used that did not match the server image thumbnails ([2aaa6de](https://github.com/RevoTale/no-js/commit/2aaa6de0b02660de34af880d484473c057223207))
* wrong short notes content disaplyed ([b8c922c](https://github.com/RevoTale/no-js/commit/b8c922c723c051a46984d932aebe3862c2db3f3a))


### Performance Improvements

* **router:** resolve slot branches in parallel during page composition ([3e4cd5e](https://github.com/RevoTale/no-js/commit/3e4cd5e4cbce93e004dec89016a8c27621e7ca87))
