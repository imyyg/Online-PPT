export const homepageContent = {
  navigation: {
    register: {
      label: '\u7acb\u5373\u6ce8\u518c',
      href: '/register',
      variant: 'primary',
      target: '_self',
      ariaLabel: '\u6ce8\u518c\u8d26\u6237',
      analytics: {
        eventName: 'cta_click',
        placement: 'header-register',
        content: '\u7acb\u5373\u6ce8\u518c'
      }
    },
    login: {
      label: '\u767b\u5f55',
      href: '/login',
      variant: 'secondary',
      target: '_self',
      ariaLabel: '\u767b\u5f55\u4f60\u7684\u8d26\u6237',
      analytics: {
        eventName: 'cta_click',
        placement: 'header-login',
        content: '\u767b\u5f55'
      }
    }
  },
  hero: {
    kicker: 'AI \u5e7b\u706f\u5e73\u53f0',
    title: '\u51e0\u5206\u949f\u5185\u5b8c\u6210AI\u6f14\u793a\u6587\u7a3f',
    subtitle: '\u4f7f\u7528\u5f15\u5bfc\u63d0\u793a\u3001\u54c1\u724c\u4e3b\u9898\u548c\u81ea\u9002\u5e94\u7248\u5e03\uff0c\u8ba9\u8981\u70b9\u8f7b\u677e\u53d8\u8eab\u7cbe\u7f8e\u6f14\u793a\u3002',
    copy: '\u51cf\u5c11\u51c6\u5907\u65f6\u95f4\uff0c\u4f7f\u7528\u9884\u7f6e\u6545\u4e8b\u7ebf\u3001\u4e0e\u5ba1\u6838\u8005\u5b9e\u65f6\u534f\u4f5c\uff0c\u65e0\u9700\u5b89\u88c5\u5373\u53ef\u5728\u4efb\u610f\u6d4f\u89c8\u5668\u4e0a\u5f00\u59cb\u6f14\u793a\u3002',
    media: {
      type: 'image',
      src: '',
      alt: '\u4ea7\u54c1\u9884\u89c8'
    },
    primaryCta: {
      label: '\u514d\u8d39\u521b\u5efa\u8d26\u6237',
      href: '/register',
      variant: 'primary',
      target: '_self',
      ariaLabel: '\u5f00\u59cb\u4f7f\u7528 Online PPT',
      analytics: {
        eventName: 'cta_click',
        placement: 'hero-primary',
        content: '\u514d\u8d39\u521b\u5efa\u8d26\u6237'
      }
    },
    secondaryCta: {
      label: '\u89c2\u770b\u6f14\u793a',
      href: '/presentations/example',
      variant: 'link',
      target: '_blank',
      ariaLabel: '\u5728\u65b0\u6807\u7b7e\u9875\u9884\u89c8\u6f14\u793a\u6587\u7a3f',
      analytics: {
        eventName: 'cta_click',
        placement: 'hero-secondary',
        content: '\u89c2\u770b\u6f14\u793a'
      }
    },
    preview: {
      heading: '\u5b9e\u65f6\u9884\u89c8',
      title: '\u667a\u80fd\u5927\u7eb2\u5df2\u751f\u6210',
      bullets: [
        'AI \u81ea\u52a8\u751f\u6210\u542b\u6838\u5fc3\u8bba\u70b9\u7684\u4e09\u5f20\u6545\u4e8b\u677f\u3002',
        '\u54c1\u724c\u8272\u677f\u81ea\u52a8\u5e94\u7528\u4e8e\u5b57\u4f53\u548c\u7126\u70b9\u5143\u7d20\u3002',
        '\u4e00\u952e\u5bfc\u51fa\u81f3\u6d4f\u89c8\u5668\u5373\u7528\u6f14\u793a\u6a21\u5f0f\u3002'
      ]
    }
  },
  features: {
    kicker: '\u56e2\u961f\u4e3a\u4f55\u9009\u62e9\u6211\u4eec',
    title: '\u9ad8\u6548\u6253\u9020\u6709\u8bf4\u670d\u529b\u7684\u6f14\u793a',
    subtitle: 'AI \u63d0\u901f\u63d0\u7a3f\u3001\u6a21\u677f\u590d\u7528\u548c\u54c1\u724c\u6cbb\u7406\uff0c\u786e\u4fdd\u6bcf\u4efd\u6f14\u793a\u51c6\u65f6\u4e0a\u7ebf\u3002',
    items: [
      {
        id: 'outline',
        icon: 'OP',
        title: '\u5f15\u5bfc\u5f0f\u5927\u7eb2\u5de5\u4f5c\u53f0',
        description: '\u4f7f\u7528\u4e13\u4e3a\u8bf4\u670d\u578b\u6f14\u793a\u8c03\u4f18\u7684AI\u63d0\u793a\uff0c\u5c06\u4f1a\u8bae\u7b14\u8bb0\u8f6c\u5316\u4e3a\u7ed3\u6784\u5316\u6545\u4e8b\u7ebf\u3002',
        bullets: [
          '\u63d0\u793a\u4e00\u952e\u751f\u6210\u5927\u7eb2\uff0c\u5728\u4e00\u5206\u949f\u5185\u5b8c\u6210\u65e5\u7a0b\u3001\u8981\u70b9\u548c\u5e7b\u706f\u76ee\u6807\u3002',
          '\u81ea\u52a8\u5e73\u8861\u7ae0\u8282\u4fe1\u606f\u5bc6\u5ea6\uff0c\u8ba9\u5ba1\u6838\u8005\u540c\u6b65\u638c\u63a7\u590d\u67e5\u7ec6\u8282\u3002',
          '\u91cd\u751f\u4f1a\u9075\u5faa\u5907\u6ce8\uff0c\u59cb\u7ec8\u4e0e\u76f8\u5173\u65b9\u53cd\u9988\u4fdd\u6301\u4e00\u81f4\u3002'
        ]
      },
      {
        id: 'templates',
        icon: 'FX',
        title: '\u9644\u5e26\u5b9e\u65f6\u9884\u89c8\u7684\u6a21\u677f\u5e93',
        description: '\u65e0\u9700\u79bb\u5f00\u7f16\u8f91\u5668\uff0c\u5373\u53ef\u6d4f\u89c8\u9500\u552e\u3001\u4ea7\u54c1\u548c\u8463\u4e8b\u4f1a\u6a21\u677f\u3002',
        bullets: [
          'AI \u57fa\u4e8e\u89d2\u8272\u3001\u53d7\u4f17\u548c\u76ee\u6807\u63a8\u8350\u6700\u9069\u5408\u7684\u6a21\u677f\u3002',
          '\u4ea4\u4e92\u5f0f\u9884\u89c8\u5c06\u524d\u4e09\u5f20\u5e7b\u706f\u4ee5\u54c1\u724c\u8272\u5f69\u5373\u65f6\u5c55\u73b0\u3002',
          '\u4e00\u952e\u5e94\u7528\u5373\u53ef\u5c06\u8349\u7a3f\u5185\u5bb9\u8f6c\u5316\u4e3a\u54c1\u724c\u89c4\u8303\u7248\u5f0f\u3002'
        ]
      },
      {
        id: 'handoff',
        icon: 'CM',
        title: '\u4e0b\u4f20\u5373\u7528\u6f14\u793a',
        description: '\u79d2\u7ea7\u4e0e\u5173\u952e\u65b9\u5171\u4eab\u4e13\u7528\u6f14\u793a\u6a21\u5f0f\u3001\u7b14\u8bb0\u548c\u5bfc\u51fa\u5305\u3002',
        bullets: [
          '\u63d0\u4f9b\u5b89\u5168\u6f14\u793a\u94fe\u63a5\uff0c\u652f\u6301\u5bc6\u7801\u548c\u8fc7\u671f\u8bbe\u7f6e\u3002',
          '\u5bfc\u51fa PDF \u6216 PowerPoint\uff0c\u4fdd\u7559\u52a8\u753b\u4e0e\u4e3b\u8bb2\u63d0\u793a\u3002',
          '\u81ea\u52a8\u53d8\u66f4\u65e5\u5fd7\u8ddf\u8e2a\u4fee\u8ba2\u4e0e\u6279\u51c6\uff0c\u4fbf\u4e8e\u5408\u89c4\u56e2\u961f\u5ba1\u67e5\u3002'
        ]
      }
    ]
  },
  socialProof: {
    kicker: '\u4f17\u591a\u56e2\u961f\u7684\u4fe1\u4efb\u4e4b\u9009',
    title: '\u8d85\u8fc77200\u4e2a\u56e2\u961f\u4f7f\u7528 Online PPT \u5feb\u901f\u4ea4\u4ed8\u6f14\u793a',
    subtitle: '\u8425\u9500\u3001\u8425\u6536\u4e0e\u4ea7\u54c1\u8d44\u6df1\u4eba\u58eb\u5747\u4fe1\u8d56\u5b9e\u65f6\u534f\u4f5c\u4e0eAI\u8f85\u52a9\u63d0\u7a3f\u3002',
    testimonials: [
      {
        quote: '\u6211\u4eec\u7684 GTM \u56e2\u961f\u73b0\u5728\u53ea\u7528\u534a\u5929\u5c31\u80fd\u628a\u6bcf\u5468\u8d4b\u80fd\u7b80\u62a5\u505a\u6210\u7cbe\u81f4\u5e7b\u706f\uff0c\u4e0d\u518d\u8017\u63892\u5929\u65f6\u95f4\u3002',
        author: 'Leah Fernandez',
        role: 'Northwind \u8d4b\u80fd\u4e3b\u7ba1'
      },
      {
        quote: 'AI \u5927\u7eb2\u548c\u54c1\u724c\u5b89\u5168\u4e3b\u9898\u8ba9\u6211\u4eec\u7684\u5206\u6790\u5e08\u4e0d\u5fc5\u6bcf\u5b63\u91cd\u5199\u5e7b\u706f\u3002',
        author: 'Marcus Lin',
        role: 'Globex Analytics \u6d1e\u5bdf\u603b\u76d1'
      },
      {
        quote: '\u5b9e\u65f6\u534f\u4f5c\u52a0\u4e0a\u6f14\u793a\u4ea4\u63a5\u6d41\u7a0b\u8ba914\u4e2a\u4ea7\u54c1\u5c0f\u7ec4\u7684\u4f1a\u8bae\u51c6\u5907\u66f4\u987a\u7545\u3002',
        author: 'Priya Desai',
        role: 'Innotech Labs \u4ea7\u54c1\u8fd0\u8425\u4e3b\u7ba1'
      }
    ],
    logos: [
      { alt: 'Northwind' },
      { alt: 'Globex Analytics' },
      { alt: 'Innotech Labs' },
      { alt: 'Acme Robotics' }
    ]
  },
  footerCta: {
    kicker: '\u6a21\u677f\u8d44\u6e90\u5e93',
    headline: '5\u5206\u949f\u5185\u642d\u5efa\u9996\u4e2aAI\u6f14\u793a',
    description: '\u6253\u5f00\u4ea4\u4e92\u5f0f\u6a21\u677f\u5305\uff0c\u4f53\u9a8c\u968f\u5f00\u968f\u7528\u7684\u6545\u4e8b\u7ebf\u548c\u54c1\u724c\u5e03\u5c40\u3002',
    primaryCta: {
      label: '\u9884\u89c8\u4ea4\u4e92\u6a21\u677f',
      href: '/presentations/example',
      variant: 'primary',
      target: '_blank',
      ariaLabel: '\u5728\u65b0\u6807\u7b7e\u9875\u6253\u5f00\u6a21\u677f\u9884\u89c8',
      analytics: {
        eventName: 'cta_click',
        placement: 'footer-primary',
        content: '\u9884\u89c8\u4ea4\u4e92\u6a21\u677f'
      }
    },
    helper: '\u65e0\u9700\u5b89\u88c5\uff0c\u76f4\u63a5\u5728\u65b0\u6807\u7b7e\u9875\u67e5\u770b\u3002'
  }
}

export function getHomepageNavigation() {
  return homepageContent.navigation
}

export function getHomepageHero() {
  return homepageContent.hero
}

export function getHomepageFeatures() {
  return homepageContent.features
}

export function getHomepageSocialProof() {
  return homepageContent.socialProof
}

export function getHomepageFooterCta() {
  return homepageContent.footerCta
}
