---
title: "Setting Up Hugo Documentation with Gismo for Our First Release"
date: 2025-01-18T10:00:00-08:00
author: "Gismo Team"
draft: true
tags: ["hugo", "documentation", "go", "release", "docsy"]
categories: ["technical", "setup"]
description: "Learn how we built our documentation site using Hugo and Docsy theme"
---

As we prepare for Gismo's first official release, we wanted to share how we built
our documentation site using Hugo and the Docsy theme. This setup has served us
well for creating comprehensive, searchable, and maintainable technical documentation.

## Why Hugo + Docsy?

When choosing our documentation platform, we had several key requirements:

- **Fast builds** for quick iteration during development
- **Professional appearance** that matches our technical audience
- **Multi-language code highlighting** for our diverse language support
- **Easy contributor workflow** using familiar Git-based editing
- **Automated deployment** with minimal maintenance overhead

Hugo with the Docsy theme checked all these boxes and more.

## Our Technical Stack

Here's what powers our documentation site:

### Core Infrastructure
- **Hugo Extended v0.148.1** - Static site generator with Sass support
- **Docsy Theme** - Google's documentation theme via Git submodule
- **GitHub Pages** - Hosting with custom domain (gismo.run)
- **GitHub Actions** - Automated build and deployment pipeline

### Development Tools
- **Node.js + npm** - For theme dependencies and build scripts
- **Dart Sass** - Advanced CSS preprocessing
- **PostCSS** - CSS optimization and modern feature support

### Content Organization
Our documentation follows a clear hierarchy:

```text
content/
├── docs/           # Main documentation
│   ├── cli/        # Command-line reference
│   ├── library/    # API documentation
│   ├── linters/    # Language-specific guides
│   └── quickstart/ # Getting started
└── blog/           # This blog section!
    └── posts/      # Individual blog posts
```

## Build and Deployment Pipeline

Our GitHub Actions workflow handles everything automatically:

1. **Trigger**: Any push to main affecting the `docs/` directory
2. **Dependencies**: Install Hugo Extended + Node.js dependencies
3. **Build**: `hugo --gc --minify` for optimized output
4. **Deploy**: Direct to GitHub Pages with custom domain

The entire process takes under 2 minutes from commit to live site.

## Key Configuration Highlights

### Hugo Configuration (hugo.toml)

```toml
baseURL = "http://gismo.run/"
title = "Gismo Documentation"
languageCode = "en-us"
theme = "docsy"

[params.docsy]
github_repo = "https://github.com/jrossi/ccfeedback"
github_branch = "main"
edit_page = true
search_enabled = true
```

### Custom Domain Setup

We use Cloudflare DNS with GitHub Pages for optimal performance:

- **CNAME record**: `gismo.run` → `your-username.github.io`
- **Static CNAME file**: `/static/CNAME` in our Hugo site
- **SSL**: Automatic via GitHub Pages + Cloudflare

## Performance Results

Our documentation site delivers excellent performance:

- **Build time**: ~15 seconds for full site rebuild
- **Page load**: Sub-second loading for most pages
- **Search**: Instant client-side search across all content
- **Mobile**: Fully responsive with touch-friendly navigation

## Content Strategy

We organize our content around user journeys:

1. **Quickstart** - Get running in 5 minutes
2. **Deep dives** - Comprehensive guides for each linter
3. **API reference** - Complete library documentation
4. **CLI reference** - Every command and flag documented

Each page includes:
- **Clear headings** with anchor links
- **Code examples** with syntax highlighting
- **Cross-references** to related sections
- **Edit links** for community contributions

## Lessons Learned

### What Worked Well

- **Git submodules** for theme management - easy updates, version control
- **Automated deployment** - zero manual steps from commit to production
- **Docsy's search** - excellent out-of-the-box search functionality
- **Hugo's speed** - rebuilds are fast enough for real-time editing

### What We'd Do Differently

- **Earlier theme customization** - we could have styled it sooner
- **Better image optimization** - compress images before committing
- **More content templates** - standardize our page structures earlier

## Looking Forward

As Gismo evolves, our documentation will grow with:

- **API examples** for all supported languages
- **Video tutorials** for complex workflows
- **Community contributions** via GitHub pull requests
- **Internationalization** as our user base expands globally

## Getting Started

Want to contribute to our documentation? Here's how:

1. **Fork** our repository on GitHub
2. **Edit** pages directly in your browser or clone locally
3. **Preview** changes with `npm run serve` in the `docs/` directory
4. **Submit** a pull request with your improvements

We welcome contributions of all sizes - from typo fixes to new sections!

---

*This post marks the beginning of our blog series leading up to Gismo's first
release. Stay tuned for more technical insights, performance deep-dives, and
community highlights.*

**Next up**: *"Go Performance Optimization Techniques We Use in Gismo"*