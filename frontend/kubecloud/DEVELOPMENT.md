# KubeCloud Frontend Development Guide

## 🚀 Getting Started

### Prerequisites
- Node.js 18+ 
- npm or yarn
- Git

### Installation
```bash
cd frontend/kubecloud
npm install
```

### Development Server
```bash
npm run dev
```

The app will be available at `http://localhost:5173`

## 📁 Project Structure

```
src/
├── components/          # Reusable Vue components
│   ├── dashboard/      # Dashboard-specific components
│   ├── features/       # Feature showcase components
│   └── ...
├── views/              # Page components
├── stores/             # Pinia state management
├── utils/              # Utility functions
├── composables/        # Vue composables
├── router/             # Vue Router configuration
├── styles/             # Global styles
└── test/               # Test setup and utilities
```

## 🛠 Development Tools

### Environment Configuration
Copy `env.example` to `.env.local` and configure:
```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:3000/api
VITE_ENABLE_MOCK_DATA=true
VITE_MOCK_DELAY=1000

# Development
VITE_ENABLE_DEBUG_MODE=true
```

### Available Scripts
- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run format` - Format code with Prettier
- `npm run test` - Run unit tests
- `npm run test:ui` - Run tests with UI
- `npm run test:coverage` - Run tests with coverage

## 🧪 Testing

### Unit Tests
Tests are written using Vitest and Vue Test Utils:

```typescript
// Example test
import { mount } from '@vue/test-utils'
import { describe, it, expect } from 'vitest'
import MyComponent from '@/components/MyComponent.vue'

describe('MyComponent', () => {
  it('renders correctly', () => {
    const wrapper = mount(MyComponent)
    expect(wrapper.text()).toContain('Expected text')
  })
})
```

### E2E Tests
E2E tests use Cypress:

```typescript
// Example E2E test
describe('Home Page', () => {
  it('should display welcome message', () => {
    cy.visit('/')
    cy.contains('Welcome to KubeCloud')
  })
})
```

## 🎨 Styling

### Design System
The app uses Vuetify 3 with a custom theme:

```typescript
// Theme colors
const theme = {
  primary: '#3B82F6',      // kubecloud blue
  secondary: '#EA580C',    // kubecloud orange
  surface: '#1E293B',      // kubecloud slate
  background: '#0F172A',   // kubecloud navy
}
```

### CSS Variables
Global CSS variables are defined in `src/assets/main.css`:

```css
:root {
  --kubecloud-blue: #3B82F6;
  --kubecloud-orange: #EA580C;
  --kubecloud-navy: #0F172A;
  --kubecloud-slate: #1E293B;
}
```

## 📊 State Management

### Pinia Stores
Stores are organized by feature:

```typescript
// Example store
export const useUserStore = defineStore('user', () => {
  const user = ref<User | null>(null)
  const isLoggedIn = computed(() => !!user.value)
  
  const login = async (credentials: LoginCredentials) => {
    // Login logic
  }
  
  return { user, isLoggedIn, login }
})
```

### Mock Data
During API development, use mock data:

```typescript
// Enable mock mode in .env.local
VITE_ENABLE_MOCK_DATA=true

// Mock data is automatically used when API is not ready
const response = await api.get('/users') // Uses mock data
```

## 🔧 Development Utilities

### Debug Tools
Development utilities are available globally:

```typescript
// Console access
window.devUtils.log.info('Debug message')
window.devUtils.performance.measure('operation', async () => {
  // Measure performance
})

// Component tracking
const tracker = componentDebug.trackComponent('MyComponent')
tracker.update(props)
tracker.unmount()
```

### Local Storage
Development storage utilities:

```typescript
devStorage.set('test-data', { key: 'value' })
const data = devStorage.get('test-data')
```

## 🚀 API Integration

### API Client
The app uses a centralized API client:

```typescript
import { api } from '@/utils/api'

// GET request
const users = await api.get('/users')

// POST request
const newUser = await api.post('/users', userData)

// With options
const response = await api.get('/users', {
  showNotifications: true,
  loadingMessage: 'Loading users...'
})
```

### Error Handling
Global error handling with notifications:

```typescript
try {
  await api.get('/users')
} catch (error) {
  // Error is automatically handled and displayed
  console.error('API Error:', error)
}
```

## 📱 Responsive Design

### Breakpoints
Vuetify breakpoints are used:

```typescript
// xs: 0-599px
// sm: 600-959px
// md: 960-1263px
// lg: 1264px+
```

### Mobile-First Approach
Components should be designed mobile-first:

```vue
<template>
  <v-container>
    <v-row>
      <v-col cols="12" md="6" lg="4">
        <!-- Content -->
      </v-col>
    </v-row>
  </v-container>
</template>
```

## 🔒 Security

### Input Validation
Use the validation utility:

```typescript
import { validateField, VALIDATION_RULES } from '@/utils/validation'

const result = validateField({
  value: email,
  rules: VALIDATION_RULES.EMAIL,
  fieldName: 'Email'
})
```

### XSS Prevention
Always sanitize user input:

```typescript
import { sanitizeInput } from '@/utils/validation'

const cleanInput = sanitizeInput(userInput)
```

## 🚀 Performance

### Code Splitting
Routes are automatically code-split:

```typescript
// Lazy-loaded component
const UserDashboard = () => import('@/views/UserDashboard.vue')
```

### Bundle Optimization
- Use dynamic imports for large components
- Optimize images and assets
- Enable gzip compression

## 🐛 Debugging

### Vue DevTools
Install Vue DevTools browser extension for debugging.

### Console Debugging
Development mode provides enhanced logging:

```typescript
// Available in development mode
devLog.info('Info message')
devLog.success('Success message')
devLog.warning('Warning message')
devLog.error('Error message')
```

### Performance Monitoring
Track component performance:

```typescript
performanceMonitor.measure('expensive-operation', async () => {
  // Expensive operation
})
```

## 📝 Code Style

### ESLint Rules
Follow the configured ESLint rules:

```bash
npm run lint
```

### Prettier Formatting
Auto-format code:

```bash
npm run format
```

### TypeScript
Use strict TypeScript:

```typescript
// Avoid 'any' type
const data: UserData = response.data

// Use proper interfaces
interface UserData {
  id: number
  name: string
  email: string
}
```

## 🚀 Deployment

### Build
```bash
npm run build
```

### Preview
```bash
npm run preview
```

### Environment Variables
Set production environment variables:

```bash
VITE_API_BASE_URL=https://api.kubecloud.com
VITE_ENABLE_MOCK_DATA=false
VITE_ENABLE_ANALYTICS=true
```

## 🤝 Contributing

### Git Workflow
1. Create feature branch
2. Make changes
3. Add tests
4. Run linting and tests
5. Submit pull request

### Commit Messages
Use conventional commits:

```
feat: add user authentication
fix: resolve navigation issue
docs: update README
test: add component tests
```

## 📚 Resources

- [Vue 3 Documentation](https://vuejs.org/)
- [Vuetify 3 Documentation](https://vuetifyjs.com/)
- [Pinia Documentation](https://pinia.vuejs.org/)
- [Vue Router Documentation](https://router.vuejs.org/)
- [Vitest Documentation](https://vitest.dev/)
- [Cypress Documentation](https://docs.cypress.io/) 