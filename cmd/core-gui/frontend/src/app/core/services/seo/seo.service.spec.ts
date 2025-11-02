import { TestBed } from '@angular/core/testing';
import { Meta, Title } from '@angular/platform-browser';
import { SeoService } from './seo.service';
import { itGood, itBad, itUgly, trio } from 'src/testing/gbu';

describe('SeoService', () => {
  let service: SeoService;
  let metaSpy: jasmine.SpyObj<Meta>;
  let titleSpy: jasmine.SpyObj<Title>;

  beforeEach(() => {
    metaSpy = jasmine.createSpyObj('Meta', ['updateTag']);
    titleSpy = jasmine.createSpyObj('Title', ['setTitle']);

    TestBed.configureTestingModule({
      providers: [
        SeoService,
        { provide: Meta, useValue: metaSpy },
        { provide: Title, useValue: titleSpy }
      ]
    });
    service = TestBed.inject(SeoService);
  });

  it('should be created', () => {
    expect(service).toBeTruthy();
  });

  describe('setMetaTitle', () => {
    trio('sets document title', {
      good: () => {
        service.setMetaTitle('Hello');
        expect(titleSpy.setTitle).toHaveBeenCalledOnceWith('Hello');
      },
      bad: () => {
        service.setMetaTitle('');
        expect(titleSpy.setTitle).toHaveBeenCalledWith('');
      },
      ugly: () => {
        // Force invalid via any cast; ensure we do not throw
        expect(() => service.setMetaTitle(null as any)).not.toThrow();
        expect(titleSpy.setTitle).toHaveBeenCalledWith(null as any);
      }
    });
  });

  describe('setMetaDescription', () => {
    itGood('updates description meta tag', () => {
      service.setMetaDescription('desc');
      expect(metaSpy.updateTag).toHaveBeenCalledWith({ name: 'description', content: 'desc' });
    });

    itBad('handles empty description', () => {
      service.setMetaDescription('');
      expect(metaSpy.updateTag).toHaveBeenCalledWith({ name: 'description', content: '' });
    });

    itUgly('does not throw on invalid description', () => {
      expect(() => service.setMetaDescription(null as any)).not.toThrow();
      expect(metaSpy.updateTag).toHaveBeenCalledWith({ name: 'description', content: null as any });
    });
  });
});
