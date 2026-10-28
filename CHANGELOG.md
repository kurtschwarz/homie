# Changelog

## [0.9.0](https://github.com/kurtschwarz/homie/compare/homie-v0.8.1...homie-v0.9.0) (2026-10-28)


### ⚠ BREAKING CHANGES

* **api:** HOMIE_WEATHERKIT_* and HOMIE_WEATHER_LAT/LON are no longer read. Weather stays off until a provider and location are added through the API.
* the tagMappings query, the saveTagMapping and deleteTagMapping mutations and the TagKind enum are replaced by calendarAssociations and create, update and deleteCalendarAssociation. An include mapping becomes an association with the family person, so an event that matches both a person and an include now shows both badges.
* **api:** HOMIE_CALDAV_* is no longer read; add accounts through the GraphQL API. The migration drops the events and tag_mappings tables, and event and tag-mapping operations now take a calendar id.
* **ui:** PersonKey and personVars are gone, Person.key is now Person.id, and Person.swatch is required. The --hm-person-* tokens and their Tailwind bridges are replaced by --hm-swatch-*.

### Features

* **agents:** add commit and prose skills ([2cb1ae9](https://github.com/kurtschwarz/homie/commit/2cb1ae98596260e0ad39b92afb7bd4e2b8420c66))
* **api:** gate the LAN surface with a bearer token ([3051ae6](https://github.com/kurtschwarz/homie/commit/3051ae6fda17920ae5aed538c2e00250b5fbcfd6))
* **api:** store calendar occurrences, people and tag mappings ([9e3de96](https://github.com/kurtschwarz/homie/commit/9e3de961ddd9060d6a50648c9699b02086097d2c))
* **api:** store weather providers in the db ([f986390](https://github.com/kurtschwarz/homie/commit/f98639035d2303d30369417044736c2143bf94e1))
* **api:** support multiple calendar accounts ([63c9862](https://github.com/kurtschwarz/homie/commit/63c9862c08993ae6222f9d1772f037627c1f91f4))
* **api:** sync the family calendar and serve it over GraphQL ([697939e](https://github.com/kurtschwarz/homie/commit/697939eeffb1b37b8691b6e1b09129100275a0dd))
* **api:** track drawings pushed to photos ([9f10419](https://github.com/kurtschwarz/homie/commit/9f104191fc2b7e4e29cb39b9d6383769d7a802e4))
* **api:** track drawings pushed to photos ([5f60ffb](https://github.com/kurtschwarz/homie/commit/5f60ffba42d5d30101da5e133523c8588d0d4ebb))
* associate people with calendars ([ef64487](https://github.com/kurtschwarz/homie/commit/ef644874387282d9705796aec024a95afe5beda7))
* **auth:** classify root fields in the schema ([132819d](https://github.com/kurtschwarz/homie/commit/132819de53947d54e59c9bcd92d0766951afd052))
* **auth:** gate the manager behind user accounts ([56c470c](https://github.com/kurtschwarz/homie/commit/56c470c162d0f21bacf15d0a923c8ffc5f916939))
* **deploy:** add uninstall script to the deb ([fb9ddc8](https://github.com/kurtschwarz/homie/commit/fb9ddc8f6dd37cf65938d2e999fa37506d1d8732))
* **deploy:** add uninstall script to the deb ([2a55e13](https://github.com/kurtschwarz/homie/commit/2a55e1371d45ed33ee0f7b0662aacbfd74bfc5a7))
* **deploy:** name the Pi homie for mDNS ([afb2d95](https://github.com/kurtschwarz/homie/commit/afb2d951bdff9f0cc9c01faa6cc009f2755dfebd))
* **deploy:** package homie-api as a .deb for the Pi ([278fd5f](https://github.com/kurtschwarz/homie/commit/278fd5f38e1ce16c3ecfcf859479058186db2572))
* **dev:** add environment reset command ([aa574ae](https://github.com/kurtschwarz/homie/commit/aa574ae110d7a755f318412bcb2102c8ffc3ef6d))
* **dev:** verify UI changes with a real Firefox MCP ([7156b96](https://github.com/kurtschwarz/homie/commit/7156b96f40689844ec253ee8ba59d4de3fbbdc00))
* **display:** add paint mode over the slideshow ([190c653](https://github.com/kurtschwarz/homie/commit/190c653cf178bcedbab2e7c4ca7494661bb01df9))
* **display:** add paint mode over the slideshow ([a3624c3](https://github.com/kurtschwarz/homie/commit/a3624c3ae040bc84bad5c45dc566dfdb567991a2))
* **display:** add schedule, backlight and wifi status ([a60b1ae](https://github.com/kurtschwarz/homie/commit/a60b1ae1d24a351d258ddc2560678c3cdfde8968))
* **display:** add schedule, backlight and wifi status ([1d67db8](https://github.com/kurtschwarz/homie/commit/1d67db88a13fa69206666b25e5c4f322223f766a))
* **display:** build the home screen ([e8d38b1](https://github.com/kurtschwarz/homie/commit/e8d38b1f9a94cf395203ac4e5031168ff449c14c))
* **display:** draw the paint canvas at 2k ([3e7fc72](https://github.com/kurtschwarz/homie/commit/3e7fc72ea54e6cbacca4215be3fd2ae3da0c62b2))
* **display:** move smooth lines into a settings popover ([65cbf50](https://github.com/kurtschwarz/homie/commit/65cbf50adfc81cd778129cf5c92828f0c26efc72))
* **display:** redesign the kiosk setup screen ([4310e8c](https://github.com/kurtschwarz/homie/commit/4310e8ca48ba6fca2e74e6dc74ebb84c7733b279))
* **display:** scale the kiosk UI to any screen size ([1618fc9](https://github.com/kurtschwarz/homie/commit/1618fc981c9c9f18c050a690b5161b8be8304cf1))
* **display:** show a ta-da dialog when saving a drawing ([#35](https://github.com/kurtschwarz/homie/issues/35)) ([733162a](https://github.com/kurtschwarz/homie/commit/733162ada4dc369752c972e7cd8bf06f9a36dc25))
* **display:** show idle mint Homie on empty canvas ([2ecc6c4](https://github.com/kurtschwarz/homie/commit/2ecc6c4deb530694acc117eab7be7bdc248acd6b))
* **display:** show the family schedule on the kiosk ([a8da37e](https://github.com/kurtschwarz/homie/commit/a8da37e4eb898275af889caf571d13606fa3951c))
* **display:** show the pairing code on the kiosk ([5678e88](https://github.com/kurtschwarz/homie/commit/5678e8813edbe4a180cfdc8c406652dfeb86c6c4))
* encode drawings at maximum jpeg quality ([7d5ecf2](https://github.com/kurtschwarz/homie/commit/7d5ecf2eab0efd626284507dd2d2be82ff491c1a))
* **events:** save mobile events to CalDAV ([9aafaf3](https://github.com/kurtschwarz/homie/commit/9aafaf3cf87c8b7c12117d131761f948fc522f4a))
* format files automatically after agent edits ([87ad814](https://github.com/kurtschwarz/homie/commit/87ad814c5976145c7ac68dac7c0ccea614fe343e))
* **ios:** add adaptive Peek app icons ([3e230dc](https://github.com/kurtschwarz/homie/commit/3e230dce26d1f43ef3ab9392ac51ecd38e8245a0))
* **ios:** add an offline mode card under no signal ([540ef87](https://github.com/kurtschwarz/homie/commit/540ef87d40c5d148e774c595e7580c91660d9e28))
* **ios:** add custom Homie pull-to-refresh ([2a13a4b](https://github.com/kurtschwarz/homie/commit/2a13a4b44c43ae6d58cfb3906e527d6032bd9913))
* **ios:** add events from the home screen ([6e85ef0](https://github.com/kurtschwarz/homie/commit/6e85ef0beacb6aa916ec0f2c01a9536a2633fc0b))
* **ios:** add home screen with kiosk preview ([1e55cab](https://github.com/kurtschwarz/homie/commit/1e55cabf74d39518178c8284c90607568029bcc8))
* **ios:** add skeleton loaders to home ([812f725](https://github.com/kurtschwarz/homie/commit/812f725d1ae378c87e93d07b0feaba5986a0e768))
* **ios:** extract HomieButton as the shared primary button ([5f14f05](https://github.com/kurtschwarz/homie/commit/5f14f052366f92082b21362f1063002993adaabd))
* **ios:** manage people from settings ([2708161](https://github.com/kurtschwarz/homie/commit/2708161bc7d463c86848ac593554dcf13bebc173))
* **ios:** refine person editor and sim build ([4b68ea5](https://github.com/kurtschwarz/homie/commit/4b68ea559e9d982f7028100c49b00e15f4a776b0))
* **ios:** reshape home dashboard ([a2d9e1b](https://github.com/kurtschwarz/homie/commit/a2d9e1b4b67864cb6d7e2d2d7b521a42f3b61d67))
* **ios:** roll Homie up to the refresh indicator ([e5e2780](https://github.com/kurtschwarz/homie/commit/e5e2780dc7677ca6732774134710fcf580c75a49))
* **ios:** show a no-signal state when Homie is unreachable ([0d1e434](https://github.com/kurtschwarz/homie/commit/0d1e4342dcddc3ef10e046821cba147f579684f8))
* **ios:** show API version and status ([78dce41](https://github.com/kurtschwarz/homie/commit/78dce415e00cfd3129047c3498f16f1a3668a768))
* **ios:** show live kitchen display on home ([e09f282](https://github.com/kurtschwarz/homie/commit/e09f28275fd409c9651ad7ca111d4e53f95dde81))
* **ios:** show offline mode indicator instead of Uh-oh title ([8073e2c](https://github.com/kurtschwarz/homie/commit/8073e2cf2fed0a6a78afaee24bb7ea62435fb259))
* **justfile:** add a clean recipe ([04df369](https://github.com/kurtschwarz/homie/commit/04df36972029b137585af264b073362aed37d17a))
* **manager:** add a sidebar admin shell ([6128272](https://github.com/kurtschwarz/homie/commit/6128272f6016cadffd61326830d7c7ae1f212cf3))
* **manager:** add ddc device picker to display page ([4ea84cf](https://github.com/kurtschwarz/homie/commit/4ea84cfc5226688b00ad6a085997ac3eb1ca3e3a))
* **manager:** add display schedule page ([3527487](https://github.com/kurtschwarz/homie/commit/3527487e5fdf9c54c11a0752e565d191bc570f10))
* **manager:** add display schedule page ([4c04c90](https://github.com/kurtschwarz/homie/commit/4c04c90adfe4b7fa21631976da3788cf8b97c4aa))
* **manager:** maintain people and tag mappings ([2ff9f47](https://github.com/kurtschwarz/homie/commit/2ff9f47aa4e95590c9e84f54916d4c6b90588b75))
* **manager:** manage multiple calendar accounts ([03b257d](https://github.com/kurtschwarz/homie/commit/03b257d05cfa84f1a04ff1c8db1df7a5e5c7146c))
* **manager:** redesign people editor ([ebff65b](https://github.com/kurtschwarz/homie/commit/ebff65b60560dc92ff554f51c0a9ff1a1e9d650d))
* **manager:** serve the manager at the site root ([193ac86](https://github.com/kurtschwarz/homie/commit/193ac8600ce03e844407773aed7a1912261e9f8b))
* **manager:** set up WeatherKit from the weather page ([64bf38d](https://github.com/kurtschwarz/homie/commit/64bf38d49cfbc21f215a5896bfb412204a93c66e))
* **people:** reorder the list by dragging ([5f3d1fc](https://github.com/kurtschwarz/homie/commit/5f3d1fc42474bde5bdcb2cf7838e1d0928794721))
* **photo-sync:** add local dev launcher ([0c89b4f](https://github.com/kurtschwarz/homie/commit/0c89b4f3521b0704c37b042840fe62d81019e891))
* **photo-sync:** add the macOS uploader app ([76cddc3](https://github.com/kurtschwarz/homie/commit/76cddc3b425cc320631659061e92dbbc83c4eaa7))
* **photo-sync:** discover Homie hosts ([a22ffde](https://github.com/kurtschwarz/homie/commit/a22ffde55724906f7fb71e517e34ba2e486df6fe))
* **photo-sync:** local photo sync development flow ([1583931](https://github.com/kurtschwarz/homie/commit/158393173fe4bc80f727dcba655a4ddd0c628110))
* **photo-sync:** release with semver tags ([4deacf2](https://github.com/kurtschwarz/homie/commit/4deacf25e93f1853874f80a654b5587343b2281b))
* **photo-sync:** release with semver tags ([bf0643b](https://github.com/kurtschwarz/homie/commit/bf0643b6f6ea2c59939c14a0819dba57bfa27f61))
* **photo-sync:** ship a notarized disk image ([26ed443](https://github.com/kurtschwarz/homie/commit/26ed443955714fe168ff5f00b9f81ffe9082c7e3))
* **photo-sync:** ship a notarized disk image ([c83e7f2](https://github.com/kurtschwarz/homie/commit/c83e7f280545ba2e8e6553f0bf6642ceb2b59b76))
* **photos:** add photo data model and GraphQL schema ([78bd089](https://github.com/kurtschwarz/homie/commit/78bd0896518d192049d300200082f5e1ccccbe1e))
* **photos:** credit a person on saved drawings ([#34](https://github.com/kurtschwarz/homie/issues/34)) ([6540ff7](https://github.com/kurtschwarz/homie/commit/6540ff72a98283c46ad99dd0a1ff61b8508a658a))
* **photos:** crop photos to a focal point ([bed145a](https://github.com/kurtschwarz/homie/commit/bed145a04a181cf2040099b98a804c6d43b20d51))
* **photos:** embed EXIF and home location in saved drawings ([#33](https://github.com/kurtschwarz/homie/issues/33)) ([0d7054a](https://github.com/kurtschwarz/homie/commit/0d7054aa852b67bf328f1d9c788473bc5c1119f5))
* **photos:** store uploads on disk and serve them ([ec0446c](https://github.com/kurtschwarz/homie/commit/ec0446c5ceb5f420e694f476e870423460d90d70))
* **photos:** track display showings ([ea16ef0](https://github.com/kurtschwarz/homie/commit/ea16ef0c931567863b56434bf53d3a7275e12bfc))
* save a drawing into the photo rotation ([378ca8b](https://github.com/kurtschwarz/homie/commit/378ca8b03f1eaf68b813dbe8242511151b90e11a))
* save a drawing into the photo rotation ([ca4bf20](https://github.com/kurtschwarz/homie/commit/ca4bf20b2b9d439f814ffcba8f47f20510dacb68))
* **timer:** add shared kitchen countdowns ([44d4ab0](https://github.com/kurtschwarz/homie/commit/44d4ab064b79f51a6fd7388155fdf930aa0bd2ba))
* **ui:** add shared design system foundations ([1979154](https://github.com/kurtschwarz/homie/commit/19791540a9608298419bebc3d98a032c3b96f197))
* **weather:** add WeatherKit widget API ([3adb7bb](https://github.com/kurtschwarz/homie/commit/3adb7bb52198f4b25f8305cc83f9af0f4f886aeb))
* **weather:** persist current conditions in SQLite ([0011637](https://github.com/kurtschwarz/homie/commit/00116376eb07367a2bb1d51b44c7a4b902607590))
* **weather:** refresh forecasts via a River queue ([9b0ef05](https://github.com/kurtschwarz/homie/commit/9b0ef05a3f9ac3cfcf11b1654d15ae7c1e76530c))


### Bug Fixes

* **api:** serve manager assets instead of index.html ([80f0fdb](https://github.com/kurtschwarz/homie/commit/80f0fdbb54c80bcd937348dfd251bf7709e4a362))
* **api:** serve the API on port 80 via socket activation ([cf1840f](https://github.com/kurtschwarz/homie/commit/cf1840f747209dd452be9d4b2477b1c438e70af9))
* **api:** serve the LAN listener from a systemd socket ([9fdd417](https://github.com/kurtschwarz/homie/commit/9fdd417d77b1788bb5793389c4c58709fb3c66d5))
* **ci:** attach release assets while releases are still drafts ([#41](https://github.com/kurtschwarz/homie/issues/41)) ([77c5cb2](https://github.com/kurtschwarz/homie/commit/77c5cb248d2969f8c183475a0314bb0b739c8cb1))
* **clients:** default Homie endpoints to port 80 ([c2fa053](https://github.com/kurtschwarz/homie/commit/c2fa0538d9226c25d8febb9a442074e308a39e8b))
* **deploy:** bind the API to port 80 via socket unit ([e5035dc](https://github.com/kurtschwarz/homie/commit/e5035dc91e3c580fe85783282c1e417de2ddfaf5))
* **deploy:** restart the kiosk after shipping a build ([8a77349](https://github.com/kurtschwarz/homie/commit/8a77349e80cfb07e463fe0403bc5a1381fbdc459))
* **deploy:** stop the API before starting its socket ([9e754bc](https://github.com/kurtschwarz/homie/commit/9e754bc3e9bfa575057bdcfa9d91dcbf6a973a72))
* **dev:** skip firefox launch if already running ([3d3373c](https://github.com/kurtschwarz/homie/commit/3d3373cadd0c46fd336699cb707aff6a357793fe))
* **dev:** use host timezone in api and firefox ([e466206](https://github.com/kurtschwarz/homie/commit/e46620687c54c610da787783183ea3e4e14936c5))
* **display:** add a kiosk page reset stylesheet ([a56281b](https://github.com/kurtschwarz/homie/commit/a56281bb2f9497a04b20b7c479906eefe1df86c9))
* **display:** keep the cursor visible in dev ([5b33eda](https://github.com/kurtschwarz/homie/commit/5b33edad524a75b734dec7c842a2cb60e25852ba))
* **display:** power the panel down for an off preview ([a822e6f](https://github.com/kurtschwarz/homie/commit/a822e6fb090b3aa3eff204018ca6e0742f1a4c3f))
* **display:** size paint panel tiles from real tokens ([c660887](https://github.com/kurtschwarz/homie/commit/c660887ac98f08467c81406f386849710afdca02))
* **ios:** force light appearance ([92e601b](https://github.com/kurtschwarz/homie/commit/92e601bcdda694c3dac6ca5df47f4d2303c60f68))
* **ios:** match home buttons to the preview radius ([d803b96](https://github.com/kurtschwarz/homie/commit/d803b96129e3862301f79fa0dfd374b4a68f0751))
* **ios:** satisfy TestFlight app validation ([c18ed82](https://github.com/kurtschwarz/homie/commit/c18ed8292a05c4f636c10e4097f819913e529e9e))
* **ios:** size Phosphor tab icons ([57256c7](https://github.com/kurtschwarz/homie/commit/57256c7ca5c97d3d3d2b184efde198ac41763e63))
* **ios:** time out graphql requests after 10 seconds ([bc0252e](https://github.com/kurtschwarz/homie/commit/bc0252e3fb3009d85b485f28244aeeddb2755b39))
* **ios:** use native sheet styling ([4749d3c](https://github.com/kurtschwarz/homie/commit/4749d3cd4a31f897d7ff2e43a5d2262bc67ace40))
* **kiosk:** blank the compositor-drawn cursor ([9dc7701](https://github.com/kurtschwarz/homie/commit/9dc77011b50d23d0434748f17b430161c9fd6e00))
* **kiosk:** prevent touch zoom and suppress prompts ([3251951](https://github.com/kurtschwarz/homie/commit/32519515cf676db19dd3d6466a2b3a59858da2f5))
* **kiosk:** render emoji in Apple's artwork ([5a50485](https://github.com/kurtschwarz/homie/commit/5a504855140711d2586c92f1e2d1fb792dda73ee))
* **kiosk:** render emoji in Apple's artwork ([c088ba6](https://github.com/kurtschwarz/homie/commit/c088ba693078937468d0866c4d880a66dec70a12))
* **ui:** settle the column before refitting the headline ([60ea3ed](https://github.com/kurtschwarz/homie/commit/60ea3edd5019c3642525a4df79c6fd1e79e549d1))


### Code Refactoring

* **ui:** colour people by swatch instead of key ([ddf573b](https://github.com/kurtschwarz/homie/commit/ddf573b11767cbb78a0bb48a3d15601a33987e6b))

## [0.8.1](https://github.com/kurtschwarz/homie/compare/homie-v0.8.0...homie-v0.8.1) (2026-10-28)


### Bug Fixes

* **ci:** attach release assets while releases are still drafts ([#41](https://github.com/kurtschwarz/homie/issues/41)) ([77c5cb2](https://github.com/kurtschwarz/homie/commit/77c5cb248d2969f8c183475a0314bb0b739c8cb1))

## [0.8.0](https://github.com/kurtschwarz/homie/compare/homie-v0.7.0...homie-v0.8.0) (2026-10-28)


### Features

* **api:** track drawings pushed to photos ([9f10419](https://github.com/kurtschwarz/homie/commit/9f104191fc2b7e4e29cb39b9d6383769d7a802e4))
* **api:** track drawings pushed to photos ([5f60ffb](https://github.com/kurtschwarz/homie/commit/5f60ffba42d5d30101da5e133523c8588d0d4ebb))
* **display:** show a ta-da dialog when saving a drawing ([#35](https://github.com/kurtschwarz/homie/issues/35)) ([733162a](https://github.com/kurtschwarz/homie/commit/733162ada4dc369752c972e7cd8bf06f9a36dc25))
* **photos:** credit a person on saved drawings ([#34](https://github.com/kurtschwarz/homie/issues/34)) ([6540ff7](https://github.com/kurtschwarz/homie/commit/6540ff72a98283c46ad99dd0a1ff61b8508a658a))
* **photos:** embed EXIF and home location in saved drawings ([#33](https://github.com/kurtschwarz/homie/issues/33)) ([0d7054a](https://github.com/kurtschwarz/homie/commit/0d7054aa852b67bf328f1d9c788473bc5c1119f5))
