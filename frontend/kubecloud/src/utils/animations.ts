// Advanced Animation Utilities for KubeCloud

export interface AnimationOptions {
  threshold?: number;
  rootMargin?: string;
  triggerOnce?: boolean;
  delay?: number;
  duration?: number;
  easing?: string;
}

export class AnimationManager {
  private observer: IntersectionObserver | null = null;
  private animatedElements: Set<Element> = new Set();

  constructor(options: AnimationOptions = {}) {
    const {
      threshold = 0.1,
      rootMargin = '0px 0px -50px 0px',
      triggerOnce = true
    } = options;

    this.observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            this.animateElement(entry.target);
            if (triggerOnce) {
              this.observer?.unobserve(entry.target);
            }
          } else if (!triggerOnce) {
            this.resetElement(entry.target);
          }
        });
      },
      {
        threshold,
        rootMargin
      }
    );
  }

  private animateElement(element: Element): void {
    if (this.animatedElements.has(element)) return;

    const delay = element.getAttribute('data-animation-delay') || '0';
    const duration = element.getAttribute('data-animation-duration') || '0.3';
    const easing = element.getAttribute('data-animation-easing') || 'ease-out';

    (element as HTMLElement).style.transition = `all ${duration}s ${easing}`;
    (element as HTMLElement).style.transitionDelay = `${delay}s`;

    // Add visible class to trigger animation
    requestAnimationFrame(() => {
      element.classList.add('visible');
      this.animatedElements.add(element);
    });
  }

  private resetElement(element: Element): void {
    element.classList.remove('visible');
    this.animatedElements.delete(element);
  }

  public observe(element: Element): void {
    this.observer?.observe(element);
  }

  public unobserve(element: Element): void {
    this.observer?.unobserve(element);
  }

  public disconnect(): void {
    this.observer?.disconnect();
  }
}

// Staggered animation utility
export function createStaggeredAnimation(
  elements: Element[],
  staggerDelay: number = 0.1,
  baseDelay: number = 0
): void {
  elements.forEach((element, index) => {
    const delay = baseDelay + (index * staggerDelay);
    element.setAttribute('data-animation-delay', delay.toString());
  });
}

// Smooth scroll utility
export function smoothScrollTo(
  target: Element | string,
  options: ScrollToOptions = {}
): void {
  const element = typeof target === 'string' ? document.querySelector(target) : target;
  if (!element) return;

  const defaultOptions = {
    behavior: 'smooth' as ScrollBehavior,
    block: 'start' as ScrollLogicalPosition,
    inline: 'nearest' as ScrollLogicalPosition
  };

  element.scrollIntoView({ ...defaultOptions, ...options });
}

// Parallax effect utility
export function createParallaxEffect(
  element: Element,
  speed: number = 0.5
): () => void {
  let ticking = false;

  const updateParallax = () => {
    const scrolled = window.pageYOffset;
    const rate = scrolled * -speed;
    (element as HTMLElement).style.transform = `translateY(${rate}px)`;
    ticking = false;
  };

  const requestTick = () => {
    if (!ticking) {
      requestAnimationFrame(updateParallax);
      ticking = true;
    }
  };

  window.addEventListener('scroll', requestTick, { passive: true });

  // Return cleanup function
  return () => {
    window.removeEventListener('scroll', requestTick);
  };
}

// Typing animation utility
export function createTypingAnimation(
  element: Element,
  text: string,
  speed: number = 50
): Promise<void> {
  return new Promise((resolve) => {
    let index = 0;
    element.textContent = '';

    const typeNextChar = () => {
      if (index < text.length) {
        element.textContent += text.charAt(index);
        index++;
        setTimeout(typeNextChar, speed);
      } else {
        resolve();
      }
    };

    typeNextChar();
  });
}

// Pulse animation utility
export function createPulseAnimation(
  element: Element,
  duration: number = 2000
): () => void {
  element.classList.add('pulse');
  
  // Return cleanup function
  return () => {
    element.classList.remove('pulse');
  };
}

// Shimmer effect utility
export function createShimmerEffect(
  element: Element,
  duration: number = 2000
): () => void {
  element.classList.add('shimmer');
  
  // Return cleanup function
  return () => {
    element.classList.remove('shimmer');
  };
}

// Bounce animation utility
export function createBounceAnimation(
  element: Element,
  duration: number = 1000
): () => void {
  element.classList.add('bounce');
  
  // Return cleanup function
  return () => {
    element.classList.remove('bounce');
  };
}

// Gradient shift animation utility
export function createGradientShift(
  element: Element,
  duration: number = 3000
): () => void {
  element.classList.add('gradient-text');
  
  // Return cleanup function
  return () => {
    element.classList.remove('gradient-text');
  };
}

// Enhanced hover effects
export function createHoverEffect(
  element: Element,
  effect: 'lift' | 'scale' | 'glow' = 'lift'
): () => void {
  element.classList.add(`hover-${effect}`);
  
  // Return cleanup function
  return () => {
    element.classList.remove(`hover-${effect}`);
  };
}

// Focus ring utility
export function createFocusRing(element: Element): () => void {
  element.classList.add('focus-ring');
  
  // Return cleanup function
  return () => {
    element.classList.remove('focus-ring');
  };
}

// Form field enhancement
export function enhanceFormField(element: Element): () => void {
  element.classList.add('form-field-enhanced');
  
  // Return cleanup function
  return () => {
    element.classList.remove('form-field-enhanced');
  };
}

// Navigation item enhancement
export function enhanceNavItem(element: Element): () => void {
  element.classList.add('nav-item-enhanced');
  
  // Return cleanup function
  return () => {
    element.classList.remove('nav-item-enhanced');
  };
}

// Glassmorphism utility
export function applyGlassmorphism(
  element: Element,
  variant: 'default' | 'light' | 'subtle' = 'default'
): () => void {
  element.classList.add(`glass${variant !== 'default' ? `-${variant}` : ''}`);
  
  // Return cleanup function
  return () => {
    element.classList.remove(`glass${variant !== 'default' ? `-${variant}` : ''}`);
  };
}

// Initialize animations on page load
export function initializeAnimations(): AnimationManager {
  const animationManager = new AnimationManager({
    threshold: 0.1,
    rootMargin: '0px 0px -50px 0px',
    triggerOnce: true
  });

  // Observe all elements with animation classes
  const animatedElements = document.querySelectorAll(
    '.fade-in, .slide-in-left, .slide-in-right, .scale-in'
  );

  animatedElements.forEach((element) => {
    animationManager.observe(element);
  });

  return animationManager;
}

// Enhanced button click effect
export function createButtonClickEffect(
  element: Element,
  effect: 'ripple' | 'scale' | 'glow' = 'ripple'
): () => void {
  const handleClick = (event: Event) => {
    const mouseEvent = event as MouseEvent;
    if (effect === 'ripple') {
      createRippleEffect(element, mouseEvent);
    } else if (effect === 'scale') {
      element.classList.add('scale-95');
      setTimeout(() => {
        element.classList.remove('scale-95');
      }, 150);
    } else if (effect === 'glow') {
      element.classList.add('hover-glow');
      setTimeout(() => {
        element.classList.remove('hover-glow');
      }, 300);
    }
  };

  element.addEventListener('click', handleClick);
  
  // Return cleanup function
  return () => {
    element.removeEventListener('click', handleClick);
  };
}

// Ripple effect utility
function createRippleEffect(element: Element, event: MouseEvent): void {
  const ripple = document.createElement('span');
  const rect = element.getBoundingClientRect();
  const size = Math.max(rect.width, rect.height);
  const x = event.clientX - rect.left - size / 2;
  const y = event.clientY - rect.top - size / 2;

  ripple.style.cssText = `
    position: absolute;
    width: ${size}px;
    height: ${size}px;
    left: ${x}px;
    top: ${y}px;
    background: rgba(255, 255, 255, 0.3);
    border-radius: 50%;
    transform: scale(0);
    animation: ripple 0.6s linear;
    pointer-events: none;
  `;

  (element as HTMLElement).style.position = 'relative';
  element.appendChild(ripple);

  setTimeout(() => {
    ripple.remove();
  }, 600);
}

// Add ripple animation to CSS if not already present
if (!document.querySelector('#ripple-animation')) {
  const style = document.createElement('style');
  style.id = 'ripple-animation';
  style.textContent = `
    @keyframes ripple {
      to {
        transform: scale(4);
        opacity: 0;
      }
    }
  `;
  document.head.appendChild(style);
}

export default {
  AnimationManager,
  createStaggeredAnimation,
  smoothScrollTo,
  createParallaxEffect,
  createTypingAnimation,
  createPulseAnimation,
  createShimmerEffect,
  createBounceAnimation,
  createGradientShift,
  createHoverEffect,
  createFocusRing,
  enhanceFormField,
  enhanceNavItem,
  applyGlassmorphism,
  initializeAnimations,
  createButtonClickEffect
}; 