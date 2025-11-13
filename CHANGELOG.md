# 📦 Changelog 
[![conventional commits](https://img.shields.io/badge/conventional%20commits-1.0.0-yellow.svg)](https://conventionalcommits.org)
[![semantic versioning](https://img.shields.io/badge/semantic%20versioning-2.0.0-green.svg)](https://semver.org)
> All notable changes to this project will be documented in this file

## [0.6.0](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.5.0...v0.6.0) (2025-11-13)

### 🍕 Features

* **cli:** set new shared dir default ([3917a94](https://github.com/ZanzyTHEbar/cursor-rules/commit/3917a94c8d01a088a8dcb52a7b53aed10807e4eb))

### 📝 Documentation

* **rules:** update internal guidance ([552447c](https://github.com/ZanzyTHEbar/cursor-rules/commit/552447c1a1ba3ac347880060c63ea53becdacf0d))

## [0.5.0](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.4.2...v0.5.0) (2025-10-31)

### 🍕 Features

* **transform:** Implement Copilot and Cursor transformers with validation and transformation logic ([8a7ff75](https://github.com/ZanzyTHEbar/cursor-rules/commit/8a7ff752b5e64b943a680fa56ef13044b612da87))
* **commands:** move commands under cursor-rules shared dir and extract generic package installer ([57869d9](https://github.com/ZanzyTHEbar/cursor-rules/commit/57869d9801da8c045f5f87c5de904af8a02257c3))

### 📝 Documentation

* enhance contributing guide and improve code quality ([5062874](https://github.com/ZanzyTHEbar/cursor-rules/commit/5062874591ec5b1bdd94b50e4fd9aff9c23e911e))

### 🧑‍💻 Code Refactoring

* **core:** consolidate symlink/stow/stub logic and standardize naming ([bb0cd51](https://github.com/ZanzyTHEbar/cursor-rules/commit/bb0cd51b9e1963f8411dbf87062fe94b23827581))
* improve code quality and respect symlink mode for cursor target ([66f4217](https://github.com/ZanzyTHEbar/cursor-rules/commit/66f4217a5a0a37a1605eea74bfdf1a7b46429e99))

### 🔁 Continuous Integration

* add security permissions and local CI testing script ([fa6c10d](https://github.com/ZanzyTHEbar/cursor-rules/commit/fa6c10dc488a4fe485cd5f6d5c0b826eecef0ce4))
* configure gosec security linter and suppress false positives ([f44e65d](https://github.com/ZanzyTHEbar/cursor-rules/commit/f44e65d1494965d9586314e0646c40d3cd561cb3)), closes [#nosec](https://github.com/ZanzyTHEbar/cursor-rules/issues/nosec)

## [0.4.2](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.4.1...v0.4.2) (2025-10-11)

### 🐛 Bug Fixes

* CI/CD linter failure: Replace deprecated filepath.HasPrefix with strings.HasPrefix ([b34d8f2](https://github.com/ZanzyTHEbar/cursor-rules/commit/b34d8f2f2ce9563523f4f7b15e2d3eba72840a60)), closes [#9](https://github.com/ZanzyTHEbar/cursor-rules/issues/9)

## [0.4.1](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.4.0...v0.4.1) (2025-08-25)

### 🐛 Bug Fixes

* address workflow duplication and Go linting errors ([e72b5f1](https://github.com/ZanzyTHEbar/cursor-rules/commit/e72b5f1395d6d326e22fe5fe691e78fccca8f1cc))

### 🔁 Continuous Integration

* fix duplicated YAML logic and create functional CI workflow ([1984b31](https://github.com/ZanzyTHEbar/cursor-rules/commit/1984b313beb91eb91cb1e65ec7ab07597d5ed0c1))
* fix workflow duplication, dependencies, and Go linting errors ([9c13ad9](https://github.com/ZanzyTHEbar/cursor-rules/commit/9c13ad987b1110115ddca5085a15f53c0171df7a)), closes [#7](https://github.com/ZanzyTHEbar/cursor-rules/issues/7)

## [0.4.0](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.3.0...v0.4.0) (2025-08-25)

### 🍕 Features

* Add nested package support with flattening as default behavior ([be42397](https://github.com/ZanzyTHEbar/cursor-rules/commit/be4239701b7c8fa14ae16628480bdf1b74cd763f)), closes [#5](https://github.com/ZanzyTHEbar/cursor-rules/issues/5)

## [0.3.0](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.2.0...v0.3.0) (2025-08-25)

### 🍕 Features

* **logger:** add go-basetools slog adapter and README notes ([ad1f558](https://github.com/ZanzyTHEbar/cursor-rules/commit/ad1f5588e1ab34a4380822904dc669f8e1b6bde2))

### 🧑‍💻 Code Refactoring

* palette architecture, AppContext, commands, tests, CI, logger improvements ([000e6d9](https://github.com/ZanzyTHEbar/cursor-rules/commit/000e6d9da57562284611353d6692f72460d51ff4))

### 🔁 Continuous Integration

* add GitHub Actions workflow for go test and golangci-lint ([d4f4cf4](https://github.com/ZanzyTHEbar/cursor-rules/commit/d4f4cf4de895e0e6fc7205b96a921c2963d7f279))

## [0.2.0](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.1.1...v0.2.0) (2025-08-22)

### 🍕 Features

* **sync:** enhance shared directory handling in sync command ([633e94c](https://github.com/ZanzyTHEbar/cursor-rules/commit/633e94c0882d7d8a30f09fd16420633134545d6b))

## [0.1.1](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.1.0...v0.1.1) (2025-08-21)

### 🐛 Bug Fixes

* normalize release asset naming ([22d00bb](https://github.com/ZanzyTHEbar/cursor-rules/commit/22d00bbcf41849f46ed9d9bf7e8991b3947eb4e8))

## [0.1.0](https://github.com/ZanzyTHEbar/cursor-rules/compare/v0.0.0...v0.1.0) (2025-08-21)

### 🍕 Features

* trigger minor release via semantic-release ([d229590](https://github.com/ZanzyTHEbar/cursor-rules/commit/d22959016b21c15edc88fea05129043c7cbc1aaf))

### 🔁 Continuous Integration

* add pnpm setup step in release workflow for improved package management ([24832f5](https://github.com/ZanzyTHEbar/cursor-rules/commit/24832f5bbd74117d6eb7697fbf2016002ad8a569))
