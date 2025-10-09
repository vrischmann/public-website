# Pico CSS Migration Plan

## Overview

This document outlines the step-by-step migration from custom CSS to Pico CSS for the public website project. The goal is to reduce custom CSS by ~40% while leveraging Pico CSS's built-in accessibility, dark mode support, and responsive design features.

## Current State Analysis

### Custom CSS Breakdown (`assets/style.css`)

#### Can be Removed/Replaced by Pico CSS (~40% reduction)
- **Basic resets**: `* { margin: 0; padding: 0; }` - Pico handles this
- **Typography**: Basic `p`, `h1-h6`, `ul`, `ol`, `li` styling
- **Links**: Basic `a` styling with hover states
- **Images**: Responsive `img` styling
- **Code blocks**: `pre` and `code` styling
- **Spacing**: Basic margin/padding patterns

#### Needs Customization (Keep but refine)
- **Font override**: `--pico-font-family-sans-serif: Times, "Times New Roman", serif;`
- **Theme system**: CSS custom properties and dark theme variables
- **Layout-specific CSS**: Container grid, navigation patterns
- **Content-specific layouts**: Blog and resume components
- **Theme toggle**: Interactive theme switching functionality

#### Can be Enhanced with Pico CSS
- **Color system**: Use Pico's semantic color variables
- **Semantic HTML**: Better use of HTML5 elements
- **Responsive design**: Pico's mobile-first approach
- **Accessibility**: Built-in focus states and ARIA support

## Migration Phases

### Phase 1: Remove Redundant Styles (Low risk)

1. **Remove Basic Typography Styles**
   - Delete lines 46-49: `* { margin: 0; padding: 0; }`
   - Delete lines 51-55: Basic `p` styling
   - Delete lines 61-64: `h1` styling
   - Delete lines 76-84: Basic list styling
   - Delete lines 86-89: `img` styling

2. **Remove Code Block Styling**
   - Delete lines 108-129: `pre` and `code` styling
   - Verify Pico's code styling works for your content

3. **Test Basic Functionality**
   - Check all pages render correctly
   - Verify typography looks good
   - Test responsive behavior

### Phase 2: Modernize with Pico CSS (Medium-high risk)

1. **Replace Custom Grid with Pico's System**
   - Use Pico's CSS custom properties for spacing
   - Replace custom `.container` grid with semantic HTML
   - Utilize Pico's responsive breakpoints

2. **Enhance Navigation with Pico Components**
   - Convert header navigation to use Pico's nav styles
   - Use Pico's button styles for theme toggle
   - Leverage Pico's link styling

3. **Update Form and Interactive Elements**
   - Use Pico's form styling if any forms exist
   - Apply Pico's focus states and hover effects
   - Test keyboard navigation accessibility

### Phase 3: Optimize Custom Layouts (High risk)

1. **Refactor Blog Layout**
   - Use Pico's card components for blog entries
   - Leverage Pico's typography scale for headings
   - Maintain custom grid structure but with Pico spacing

2. **Modernize Resume Layout**
   - Use Pico's grid system where appropriate
   - Apply Pico's color variables for better theming
   - Keep custom resume-specific layouts but clean up

3. **Enhance Theme Integration**
   - Integrate custom theme variables with Pico's system
   - Ensure dark mode works seamlessly with Pico
   - Test theme toggle functionality

### Phase 4: Final Polish and Testing

1. **Cross-browser Testing**
   - Test on Chrome, Firefox, Safari, Edge
   - Verify mobile responsiveness
   - Check accessibility with screen readers

2. **Performance Optimization**
   - Remove unused CSS
   - Minimize final CSS files
   - Test loading performance

3. **Documentation Update**
   - Update CLAUDE.md with new CSS structure
   - Document any new Pico CSS usage patterns
   - Note any deviations from Pico defaults

## Detailed Implementation Steps

### Step 1: Essential Styles to Keep in `style.css`

```css
/* Keep these essential styles in style.css */
:root {
  --background-color: #fff;
  --text-color: #000;
  --text-secondary: #667;
  --link-color: blue;
  --header-border-color: #000;
  --code-background-color: #eee;
  --code-text-color: #ec0000;
  --accent-color: #d21c1c;
  --bg-gray-focus: rgba(0, 0, 0, 0.1);

  /* Font override for Pico CSS */
  --pico-font-family-sans-serif: Times, "Times New Roman", serif;
}

[data-theme='dark'] {
  /* Keep dark theme variables */
  /* ... existing dark theme variables ... */
}

/* Theme toggle functionality */
#theme-toggle {
  /* Keep theme toggle styles */
  /* ... existing theme toggle styles ... */
}
```

### Step 2: HTML Template Updates

Update templates to use more semantic HTML where possible:

```html
<!-- Before -->
<div class="container">
  <header>
    <ul>
      <li><a href="/">Home</a></li>
    </ul>
  </header>
</div>

<!-- After (with Pico CSS) -->
<div class="container">
  <nav>
    <ul>
      <li><a href="/">Home</a></li>
    </ul>
  </nav>
</div>
```

## Risk Assessment

### Low Risk Changes
- Removing basic typography resets
- Using Pico's color variables
- Basic theme integration
- Creating backup (optional but recommended)

### Medium Risk Changes
- Navigation styling changes
- Form element updates
- Responsive breakpoint adjustments

### High Risk Changes
- Custom layout grid modifications
- Blog and resume component restructuring
- Major theme system changes

## Testing Strategy

### Automated Testing
- Visual regression testing with screenshots
- CSS validation with W3C validator
- Performance testing with Lighthouse

### Manual Testing Checklist
- [ ] All pages render correctly
- [ ] Typography is readable and well-spaced
- [ ] Navigation works on all screen sizes
- [ ] Theme toggle functions properly
- [ ] Dark mode looks good
- [ ] Blog layout maintains functionality
- [ ] Resume layout preserves structure
- [ ] Code blocks display correctly
- [ ] Images are responsive
- [ ] Links have proper hover states
- [ ] Keyboard navigation works
- [ ] Mobile responsiveness verified

## Rollback Plan

If issues arise during migration:
1. Revert to `assets/style.css.backup`
2. Remove `assets/custom.css`
3. Test functionality is restored
4. Address specific issues before retrying

## Success Metrics

- **CSS Reduction**: Custom CSS reduced by 40% (from 497 lines to ~300 lines)
- **Performance**: Page load time improves by removing redundant CSS
- **Accessibility**: WCAG compliance maintained or improved
- **Maintainability**: Easier to modify and extend styles
- **User Experience**: Visual design quality maintained or enhanced

## Timeline Estimates

- **Phase 1**: 1 hour (remove redundant styles, test)
- **Phase 2**: 2 hours (modernize with Pico CSS)
- **Phase 3**: 2 hours (optimize custom layouts)
- **Phase 4**: 1 hour (final testing and polish)

**Total Estimated Time**: 6 hours

## Next Steps

1. Review this plan and adjust based on project priorities
2. Start with Phase 1 (remove redundant styles)
3. Test thoroughly after each phase
4. Document any deviations from this plan
5. Update project documentation upon completion

---

*Last Updated: 2025-10-09*
*Author: Claude Code Assistant*