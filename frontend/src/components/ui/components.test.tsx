import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import React from 'react';
import {
  Button,
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
  Badge,
  Skeleton,
  EmptyState,
  ErrorState,
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
  Tabs,
  TabsList,
  TabsTrigger,
  TabsContent,
} from './index';

describe('UI Components', () => {
  describe('Button', () => {
    it('renders text and handles click', () => {
      const handleClick = vi.fn();
      render(<Button onClick={handleClick}>Click Me</Button>);
      const btn = screen.getByRole('button', { name: /click me/i });
      fireEvent.click(btn);
      expect(handleClick).toHaveBeenCalledTimes(1);
    });

    it('disables interactions when disabled or loading', () => {
      const handleClick = vi.fn();
      const { rerender } = render(
        <Button disabled onClick={handleClick}>
          Disabled
        </Button>
      );
      const btn = screen.getByRole('button');
      expect(btn).toBeDisabled();
      fireEvent.click(btn);
      expect(handleClick).not.toHaveBeenCalled();

      rerender(
        <Button isLoading onClick={handleClick}>
          Loading
        </Button>
      );
      expect(btn).toBeDisabled();
    });
  });

  describe('Card', () => {
    it('renders complete card structure', () => {
      render(
        <Card>
          <CardHeader>
            <CardTitle>Security Status</CardTitle>
            <CardDescription>Device scan summary</CardDescription>
          </CardHeader>
          <CardContent>No active threats</CardContent>
          <CardFooter>Updated just now</CardFooter>
        </Card>
      );
      expect(screen.getByText('Security Status')).toBeInTheDocument();
      expect(screen.getByText('Device scan summary')).toBeInTheDocument();
      expect(screen.getByText('No active threats')).toBeInTheDocument();
      expect(screen.getByText('Updated just now')).toBeInTheDocument();
    });
  });

  describe('Badge', () => {
    it('renders critical severity with label and count', () => {
      render(<Badge severity="critical" count={3} />);
      expect(screen.getByText('Critical')).toBeInTheDocument();
      expect(screen.getByText('3')).toBeInTheDocument();
    });

    it('renders all severities with their labels', () => {
      const { rerender } = render(<Badge severity="high" />);
      expect(screen.getByText('High')).toBeInTheDocument();

      rerender(<Badge severity="medium" />);
      expect(screen.getByText('Medium')).toBeInTheDocument();

      rerender(<Badge severity="low" />);
      expect(screen.getByText('Low')).toBeInTheDocument();

      rerender(<Badge severity="info" />);
      expect(screen.getByText('Info')).toBeInTheDocument();

      rerender(<Badge severity="ok" />);
      expect(screen.getByText('Passed')).toBeInTheDocument();
    });
  });

  describe('Skeleton', () => {
    it('renders with pulse animation class', () => {
      const { container } = render(<Skeleton className="w-20 h-4" />);
      expect(container.firstChild).toHaveClass('animate-pulse');
    });
  });

  describe('EmptyState', () => {
    it('renders title, description, and action button', () => {
      render(
        <EmptyState
          title="No Applications Found"
          description="Try adjusting your search criteria."
          action={<Button size="sm">Reset Filter</Button>}
        />
      );
      expect(screen.getByText('No Applications Found')).toBeInTheDocument();
      expect(screen.getByText('Try adjusting your search criteria.')).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /reset filter/i })).toBeInTheDocument();
    });
  });

  describe('ErrorState', () => {
    it('renders standard error code, description, and triggers retry', () => {
      const handleRetry = vi.fn();
      render(
        <ErrorState
          code="AdbMissing"
          title="ADB Binary Not Found"
          remedy="Please check your system path or download platform-tools."
          onRetry={handleRetry}
        />
      );
      expect(screen.getByText('AdbMissing')).toBeInTheDocument();
      expect(screen.getByText('ADB Binary Not Found')).toBeInTheDocument();
      expect(screen.getByText('Please check your system path or download platform-tools.')).toBeInTheDocument();

      const retryBtn = screen.getByRole('button', { name: /retry/i });
      fireEvent.click(retryBtn);
      expect(handleRetry).toHaveBeenCalledTimes(1);
    });
  });

  describe('Table', () => {
    it('renders desktop density table rows and cells', () => {
      render(
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Package</TableHead>
              <TableHead>Version</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow>
              <TableCell>com.android.settings</TableCell>
              <TableCell>12.0</TableCell>
            </TableRow>
          </TableBody>
        </Table>
      );
      expect(screen.getByText('Package')).toBeInTheDocument();
      expect(screen.getByText('com.android.settings')).toBeInTheDocument();
    });
  });

  describe('Tabs', () => {
    it('switches tab content on trigger click', () => {
      render(
        <Tabs defaultValue="tab1">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
            <TabsTrigger value="tab2">Tab 2</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Content 1</TabsContent>
          <TabsContent value="tab2">Content 2</TabsContent>
        </Tabs>
      );
      expect(screen.getByText('Content 1')).toBeInTheDocument();
      const tab2Trigger = screen.getByRole('tab', { name: /tab 2/i });
      fireEvent.keyDown(tab2Trigger, { key: 'Enter', code: 'Enter' });
      expect(screen.getByText('Content 2')).toBeInTheDocument();
    });
  });
});
