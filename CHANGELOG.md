# Changelog

All notable changes to this project will be documented in this file.

## [1.1.6](https://github.com/microsoft/kiota-serialization-multipart-go/compare/v1.1.5...v1.1.6) (2026-11-11)


### Bug Fixes

* **multipart:** handle nil and empty collections in WriteCollection methods ([c15b8ca](https://github.com/microsoft/kiota-serialization-multipart-go/commit/c15b8ca48e213d6e928a6e4aa3384504c9f06ad0))
* **multipart:** normalize header keys and prevent malformed headers in WriteStringValue ([a2939ce](https://github.com/microsoft/kiota-serialization-multipart-go/commit/a2939ce4c48f6192a035bb53dddec74317465faa))
* **multipart:** normalize header keys and validate CRLF in WriteStringValue ([846e008](https://github.com/microsoft/kiota-serialization-multipart-go/commit/846e00845549c48f18bf9f5abe7bc57388694a28))
* **multipart:** omit nil header string values and handle nil buffer in get serialized content ([32c1695](https://github.com/microsoft/kiota-serialization-multipart-go/commit/32c16958ef493cd368db61a5cb9808d761b4c467))
* **multipart:** reject CRLF in header keys prior to whitespace trimming ([49b81cb](https://github.com/microsoft/kiota-serialization-multipart-go/commit/49b81cb3557bcdd944f9ac171c751e738357e1f6))
* **multipart:** validate header keys and values against CRLF injection in WriteStringValue ([0400eee](https://github.com/microsoft/kiota-serialization-multipart-go/commit/0400eeedd207ff97c278d14245fec932d8850333))

## [1.1.5](https://github.com/microsoft/kiota-serialization-multipart-go/compare/v1.1.4...v1.1.5) (2026-11-03)


### Bug Fixes

* **factory:** add nil receiver check to GetValidContentType ([65ff36e](https://github.com/microsoft/kiota-serialization-multipart-go/commit/65ff36e01870558306038278544e4d7d8386e41c))
* **factory:** add nil receiver check to GetValidContentType ([c9c5163](https://github.com/microsoft/kiota-serialization-multipart-go/commit/c9c516321208acd5638ea91916ee159a93ee3c51))
* **multipart:** add defensive nil receiver guards to serialization callback methods ([68233da](https://github.com/microsoft/kiota-serialization-multipart-go/commit/68233dace8576e9ecec7951871ec0b5b563c63b4))
* **multipart:** add nil guards to unsupported methods and validate multipart body before start callback ([1df8d43](https://github.com/microsoft/kiota-serialization-multipart-go/commit/1df8d4357a50847d013be62fb943af4eeaf676ce))
* **multipart:** add nil guards to unsupported methods and validate multipart body before start callback ([2cf9b02](https://github.com/microsoft/kiota-serialization-multipart-go/commit/2cf9b027f1ad113a5cb9fbccd416265fd6e103d1))
* **multipart:** guard callback getters and setters against nil pointer dereference ([4db85ef](https://github.com/microsoft/kiota-serialization-multipart-go/commit/4db85efd1a13bd83dfc61156a5ccacffff2a1d2e))
* **multipart:** handle nil and empty values in WriteAdditionalData and WriteAnyValue ([1ac92d7](https://github.com/microsoft/kiota-serialization-multipart-go/commit/1ac92d77da38042b9ed854d64c7caf9b088cd55c))
* **multipart:** handle nil and empty values in WriteAdditionalData and WriteAnyValue ([b6634c3](https://github.com/microsoft/kiota-serialization-multipart-go/commit/b6634c3a039b389e48585472fb061f97fbfd92b8))
* **multipart:** handle nil values in scalar write methods and WriteStringValue ([0c282a4](https://github.com/microsoft/kiota-serialization-multipart-go/commit/0c282a4d52740e5d923c39f882d21ff5b6e6fd5a))
* **multipart:** handle nil values in scalar write methods and WriteStringValue ([431a897](https://github.com/microsoft/kiota-serialization-multipart-go/commit/431a8977bbb8511972eb6bb01b23ab58e9ff0758))

## [1.1.4](https://github.com/microsoft/kiota-serialization-multipart-go/compare/v1.1.3...v1.1.4) (2026-10-21)


### Bug Fixes

* **multipart:** return potential errors in WriteStringValue ([669d99e](https://github.com/microsoft/kiota-serialization-multipart-go/commit/669d99ec1a16221dad9abbc8ca935831a1bca783))
* **multipart:** trim whitespace on content type check and add nil receiver guards ([0df2716](https://github.com/microsoft/kiota-serialization-multipart-go/commit/0df2716e6db2e4d0c18248c74f57bc5ecaf11798))
* **multipart:** trim whitespace on content type check and add nil receiver guards ([a13ef4f](https://github.com/microsoft/kiota-serialization-multipart-go/commit/a13ef4f9429868c49c43177befd035ae474470b7))

## [1.1.3](https://github.com/microsoft/kiota-serialization-multipart-go/compare/v1.1.2...v1.1.3) (2026-10-20)


### Bug Fixes

* **multipart:** accept whitespace before parameters ([f489f05](https://github.com/microsoft/kiota-serialization-multipart-go/commit/f489f05bdfbb64d04ba3a7b12aff894f66663cf2))
* **multipart:** support content type parameters and check serialization error before callback ([96fe37e](https://github.com/microsoft/kiota-serialization-multipart-go/commit/96fe37e98c6acd42819de1baf1df5f4b5cd7aed6))
* **multipart:** support content type parameters and check serialization error before callback ([c316d68](https://github.com/microsoft/kiota-serialization-multipart-go/commit/c316d68ee2be06e5ef0e406eb7e94e79f7025d9e))

## [1.1.2](https://github.com/microsoft/kiota-serialization-multipart-go/compare/v1.1.1...v1.1.2) (2025-04-02)


### Bug Fixes

* removes common go dependency ([d4086ca](https://github.com/microsoft/kiota-serialization-multipart-go/commit/d4086caae32a827aba998ffb0f68f2da912b87e8))
* removes common go dependency ([028bb1e](https://github.com/microsoft/kiota-serialization-multipart-go/commit/028bb1e66ebaae555eb85beac967f23116722119))

## [1.1.1](https://github.com/microsoft/kiota-serialization-multipart-go/compare/v1.1.0...v1.1.1) (2025-03-24)


### Bug Fixes

* upgrades common go dependency to solve triming issues ([202dba2](https://github.com/microsoft/kiota-serialization-multipart-go/commit/202dba266af41968af8d61306a436ec0fdc69b38))
* upgrades common go dependency to solve triming issues ([aba6ac5](https://github.com/microsoft/kiota-serialization-multipart-go/commit/aba6ac56bcdb6dd8715eb1c09b3c28c0530f4f4f))

## [1.1.0](https://github.com/microsoft/kiota-serialization-multipart-go/compare/v1.0.0...v1.1.0) (2025-03-13)


### Features

* upgrades required go version from go1.18 to go 1.22 ([66b4e94](https://github.com/microsoft/kiota-serialization-multipart-go/commit/66b4e94f0476a5ddc25d021d0d15128bb297f351))

## [1.0.0] - 2023-07-26

### Changed

- Initial release
